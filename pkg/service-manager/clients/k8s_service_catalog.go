///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package clients

// NO TEST

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/vmware/dispatch/pkg/client"

	"github.com/pkg/errors"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	dispatchv1 "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
)

// K8sBrokerConfigOpts are k8s specific configuratation options
type K8sBrokerConfigOpts struct {
	K8sConfig        string
	CatalogNamespace string
	SecretStoreURL   string
}

type k8sServiceCatalogClient struct {
	clientset     *kubernetes.Clientset
	sdk           *servicecatalog.SDK
	config        K8sBrokerConfigOpts
	secretsClient client.SecretsClient
}

// NewK8sBrokerClient creates a new K8s-backed Service Broker
func NewK8sBrokerClient(config K8sBrokerConfigOpts) (BrokerClient, error) {

	var err error
	var k8sConfig *rest.Config
	if config.K8sConfig == "" {
		// creates the in-cluster config
		k8sConfig, err = rest.InClusterConfig()
	} else {
		// create from a configuration
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", config.K8sConfig)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes config")
	}

	c, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes clientset")
	}

	sc, err := clientset.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes service catalog clientset")
	}

	sdk := &servicecatalog.SDK{ServiceCatalogClient: sc}

	return &k8sServiceCatalogClient{
		clientset:     c,
		sdk:           sdk,
		config:        config,
		secretsClient: client.NewSecretsClient(config.SecretStoreURL, client.AuthWithToken("cookie"), ""),
	}, nil
}

// ListServiceClasses returns a list of ServiceClass entities which correspond to the available services.  Each
// service includes the avaialable plans
func (c *k8sServiceCatalogClient) ListServiceClasses() ([]entitystore.Entity, error) {

	cscs, err := c.sdk.RetrieveClasses()
	if err != nil {
		return nil, errors.Wrapf(err, "Error retreiving service class list")
	}
	var serviceClasses []entitystore.Entity

	for _, csc := range cscs {
		log.Debugf("Syncing service class %s", csc.Spec.ExternalName)
		if csc.Status.RemovedFromBrokerCatalog {
			continue
		}
		plans, err := c.sdk.RetrievePlansByClass(&csc)
		if err != nil {
			return nil, errors.Wrapf(err, "Error retreiving plan for service class %s", csc.Spec.ExternalName)
		}
		log.Debugf("Fetch plans for %s [%d plans]", csc.Spec.ExternalName, len(plans))
		var serviceClassPlans []entities.ServicePlan
		for _, plan := range plans {
			log.Debugf("Parsing plan %s", plan.Spec.ExternalName)
			if plan.Status.RemovedFromBrokerCatalog {
				log.Debugf("Plan %s removed from broker catalog", plan.Spec.ExternalName)
				continue
			}
			marshall := func(ps *runtime.RawExtension) (*spec.Schema, error) {
				out := new(spec.Schema)
				if ps != nil && ps.Size() > 0 {
					b, _ := json.Marshal(ps)
					if err := json.Unmarshal(b, out); err != nil {
						return nil, errors.Wrap(err, "could not decode schema")
					}
				}
				return out, nil
			}
			create, err := marshall(plan.Spec.ServiceInstanceCreateParameterSchema)
			if err != nil {
				log.Debugf("Plan %s failed to marshall create schema", plan.Spec.ExternalName)
				return nil, errors.Wrap(err, "Error marshalling create schema")
			}
			update, err := marshall(plan.Spec.ServiceInstanceUpdateParameterSchema)
			if err != nil {
				log.Debugf("Plan %s failed to marshall update schema", plan.Spec.ExternalName)
				return nil, errors.Wrap(err, "Error marshalling update schema")
			}
			bind, err := marshall(plan.Spec.ServiceBindingCreateParameterSchema)
			if err != nil {
				log.Debugf("Plan %s failed to marshall bind schema", plan.Spec.ExternalName)
				return nil, errors.Wrap(err, "Error marshalling bind schema")
			}

			bindable := csc.Spec.Bindable
			if plan.Spec.Bindable != nil {
				bindable = *plan.Spec.Bindable
			}

			serviceClassPlans = append(serviceClassPlans, entities.ServicePlan{
				BaseEntity: entitystore.BaseEntity{
					Name:   plan.Spec.ExternalName,
					Status: entitystore.StatusREADY,
				},
				PlanID:      plan.Name,
				Description: plan.Spec.Description,
				Schema: entities.Schema{
					Create: create,
					Update: update,
					Bind:   bind,
				},
				Free:     plan.Spec.Free,
				Bindable: bindable,
				Metadata: plan.Spec.ExternalMetadata,
			})
		}
		serviceClasses = append(serviceClasses, &entities.ServiceClass{
			BaseEntity: entitystore.BaseEntity{
				Name:   csc.Spec.ExternalName,
				Status: entitystore.StatusREADY,
			},
			ServiceID: csc.Name,
			Broker:    csc.Spec.ClusterServiceBrokerName,
			Bindable:  csc.Spec.Bindable,
			Plans:     serviceClassPlans,
		})
	}

	return serviceClasses, nil
}

// ListServiceInstances returns a list of ServiceInstance entities which correspond to the provisioned services.  Most
// importantly the status of each service instance is returned as well.
func (c *k8sServiceCatalogClient) ListServiceInstances() ([]entitystore.Entity, error) {
	instances, err := c.sdk.RetrieveInstances(c.config.CatalogNamespace)
	if err != nil {
		return nil, errors.Wrapf(err, "Error retreiving service instance list")
	}
	var serviceInstances []entitystore.Entity
	for _, instance := range instances.Items {

		parameters := make(map[string]interface{})
		if instance.Spec.Parameters != nil {
			err := json.Unmarshal(instance.Spec.Parameters.Raw, &parameters)
			if err != nil {
				log.Errorf("Failed to decode the service instance parameters for %s", instance.ObjectMeta.Name)
			}
		}
		orgID, ok := instance.Labels["org"]
		if !ok {
			// Log the error and continue
			log.Errorf("Error no org for service instance %s", instance.Name)
			continue
		}
		serviceInstance := &entities.ServiceInstance{
			BaseEntity: entitystore.BaseEntity{
				OrganizationID: orgID,
				ID:             instance.ObjectMeta.Name,
				Status:         entitystore.StatusUNKNOWN,
			},
			ServiceClass: instance.Spec.ClusterServiceClassExternalName,
			ServicePlan:  instance.Spec.ClusterServicePlanExternalName,
			Parameters:   parameters,
		}
		for _, cond := range instance.Status.Conditions {
			if cond.Type == v1beta1.ServiceInstanceConditionReady && cond.Status == v1beta1.ConditionTrue {
				serviceInstance.Status = entitystore.StatusREADY
				break
			}
			if cond.Type == v1beta1.ServiceInstanceConditionReady && cond.Status == v1beta1.ConditionFalse {
				// This condition is returned when provisioning or a non-existent plan is referenced
				if cond.Reason == "Provisioning" {
					continue
				}
				serviceInstance.Status = entitystore.StatusERROR
				log.Debugf("Recording service instance error: %s", cond.Message)
				serviceInstance.Reason = append(serviceInstance.Reason, cond.Message)
				break
			}
			if cond.Type == v1beta1.ServiceInstanceConditionFailed && cond.Status == v1beta1.ConditionTrue {
				serviceInstance.Status = entitystore.StatusERROR
				log.Debugf("Recording service instance error: %s", cond.Message)
				serviceInstance.Reason = append(serviceInstance.Reason, cond.Message)
				break
			}
		}
		serviceInstances = append(serviceInstances, serviceInstance)
	}
	return serviceInstances, nil
}

// ListServiceBindings returns a list of ServiceBinding entities which correspond to the service bindings.  The
// bindings are separate entities and have a status of their own.
func (c *k8sServiceCatalogClient) ListServiceBindings() ([]entitystore.Entity, error) {
	instances, err := c.sdk.RetrieveInstances(c.config.CatalogNamespace)
	if err != nil {
		return nil, errors.Wrapf(err, "Error retreiving service instance list")
	}
	instanceMap := make(map[string]*v1beta1.ServiceInstance)
	for _, instance := range instances.Items {
		instanceMap[instance.Name] = instance.DeepCopy()
	}
	bindings, err := c.sdk.RetrieveBindings(c.config.CatalogNamespace)
	var serviceBindings []entitystore.Entity
	for _, binding := range bindings.Items {
		instance := instanceMap[binding.Spec.ServiceInstanceRef.Name]

		parameters := make(map[string]interface{})
		if binding.Spec.Parameters != nil {
			err := json.Unmarshal(binding.Spec.Parameters.Raw, &parameters)
			if err != nil {
				log.Errorf("Failed to decode the service binding parameters for %s", instance.Name)
			}
		}
		serviceBinding := &entities.ServiceBinding{
			BaseEntity: entitystore.BaseEntity{
				Status: entitystore.StatusUNKNOWN,
			},
			Parameters: parameters,
			BindingID:  binding.Name,
		}
		for _, cond := range binding.Status.Conditions {
			if cond.Type == v1beta1.ServiceBindingConditionReady && cond.Status == v1beta1.ConditionTrue {
				secretsClient := c.clientset.CoreV1().Secrets(c.config.CatalogNamespace)
				boundSecret, err := secretsClient.Get(binding.Spec.SecretName, metav1.GetOptions{})
				if err != nil {
					// Should probably not fail here... allow retry
					log.Errorf("Error fetching secret %s from kubernetes: %v", binding.Spec.SecretName, err)
					continue
				}
				// This is rerun/created every single loop... should only do once or on change
				secrets := make(map[string]string)
				for k, v := range boundSecret.Data {
					secrets[k] = string(v)
				}
				// This is a hack... we shouldn't be setting secrets on a list method
				orgID, ok := binding.Labels["org"]
				if !ok {
					// Log the error and continue
					log.Errorf("Error no org for binding %s", binding.Name)
					continue
				}
				err = c.setSecret(orgID, binding.Name, secrets)
				if err != nil {
					// Log the error and continue
					log.Errorf("Error setting secret for binding %s: %v", binding.Name, err)
					continue
				}

				serviceBinding.OrganizationID = orgID
				serviceBinding.Status = entitystore.StatusREADY
				break
			}
			if cond.Type == v1beta1.ServiceBindingConditionFailed && cond.Status == v1beta1.ConditionTrue {
				serviceBinding.Status = entitystore.StatusERROR
				log.Debugf("Recording service binding error: %s", cond.Message)
				serviceBinding.Reason = append(serviceBinding.Reason, cond.Message)
				break
			}
		}
		log.Debugf("Adding binding %s", serviceBinding.BindingID)
		serviceBindings = append(serviceBindings, serviceBinding)
	}
	return serviceBindings, nil
}

// CreateService provisions a service, creating a service instance.
func (c *k8sServiceCatalogClient) CreateService(class *entities.ServiceClass, service *entities.ServiceInstance) error {
	secrets, err := c.getSecrets(service.OrganizationID, service.SecretParameters)
	if err != nil {
		service.SetStatus(entitystore.StatusERROR)
		return errors.Wrapf(err, "Error fetching secrets for provisioning service %s of class %s with plan %s", service.Name, service.ServiceClass, service.ServicePlan)
	}

	serviceParamsJSON, _ := json.Marshal(service.Parameters)

	instance, err := c.sdk.ServiceCatalog().ServiceInstances(c.config.CatalogNamespace).Create(&v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ID,
			Namespace: c.config.CatalogNamespace,
			Labels:    map[string]string{"org": service.OrganizationID},
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: service.ServiceClass,
				ClusterServicePlanExternalName:  service.ServicePlan,
			},
			Parameters:     &runtime.RawExtension{Raw: serviceParamsJSON},
			ParametersFrom: servicecatalog.BuildParametersFrom(secrets),
		},
	})

	if err != nil {
		service.SetStatus(entitystore.StatusERROR)
		return errors.Wrapf(err, "Error provisioning service %s of class %s with plan %s", service.Name, service.ServiceClass, service.ServicePlan)
	}
	service.SetStatus(entitystore.StatusCREATING)
	service.InstanceID = instance.Name
	return nil
}

// CreateBinding creates a binding (credentials) for a service.
func (c *k8sServiceCatalogClient) CreateBinding(service *entities.ServiceInstance, binding *entities.ServiceBinding) error {
	log.Debugf("Creating service binding for service %+v and binding %+v", service, binding)
	secrets, err := c.getSecrets(binding.OrganizationID, binding.SecretParameters)
	if err != nil {
		binding.SetStatus(entitystore.StatusERROR)
		return errors.Wrapf(err, "Error fetching secrets for binding service %s of class %s with plan %s", service.Name, service.ServiceClass, service.ServicePlan)
	}

	bindingParamsJSON, _ := json.Marshal(binding.Parameters)

	b, err := c.sdk.ServiceCatalog().ServiceBindings(c.config.CatalogNamespace).Create(&v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ID,
			Namespace: c.config.CatalogNamespace,
			Labels:    map[string]string{"org": service.OrganizationID},
		},
		Spec: v1beta1.ServiceBindingSpec{
			ServiceInstanceRef: v1beta1.LocalObjectReference{
				Name: service.ID,
			},
			SecretName:     binding.BindingSecret,
			Parameters:     &runtime.RawExtension{Raw: bindingParamsJSON},
			ParametersFrom: servicecatalog.BuildParametersFrom(secrets),
		},
	})

	if err != nil {
		binding.SetStatus(entitystore.StatusERROR)
		return errors.Wrapf(err, "Error binding service %s of class %s with plan %s", service.Name, service.ServiceClass, service.ServicePlan)
	}
	binding.SetStatus(entitystore.StatusCREATING)
	binding.BindingID = b.Name
	return nil
}

// DeleteService deprovisions a service.
func (c *k8sServiceCatalogClient) DeleteService(service *entities.ServiceInstance) error {
	err := c.sdk.Deprovision(c.config.CatalogNamespace, service.ID)
	if err != nil {
		return errors.Wrapf(err, "Error deleting service instance %s", service.Name)
	}
	service.Status = entitystore.StatusDELETED
	return nil
}

func (c *k8sServiceCatalogClient) DeleteBinding(binding *entities.ServiceBinding) error {
	err := c.sdk.Unbind(c.config.CatalogNamespace, binding.BindingID)
	if err != nil {
		// Nothing we can do... try again later if there are orphaned resources
		log.Errorf("Error deleting service binding %s", binding.BindingID)
	}
	err = c.deleteSecret(binding.OrganizationID, binding.BindingID)
	if err != nil {
		return err
	}
	binding.Status = entitystore.StatusDELETED
	return nil
}

func (c *k8sServiceCatalogClient) getSecrets(organizationID string, secretNames []string) (map[string]string, error) {
	secrets := make(map[string]string)
	for _, name := range secretNames {
		resp, err := c.secretsClient.GetSecret(context.Background(), organizationID, name)
		if err != nil {
			return secrets, errors.Wrapf(err, "failed to get secrets from secret store")
		}
		for key, value := range resp.Secrets {
			secrets[key] = value
		}
	}
	return secrets, nil
}

func (c *k8sServiceCatalogClient) deleteSecret(organizationID string, secretName string) error {
	err := c.secretsClient.DeleteSecret(context.Background(), organizationID, secretName)
	if err != nil {
		_, ok := err.(*secret.GetSecretNotFound)
		if !ok {
			return errors.Wrapf(err, "failed to delete secret %s for binding", secretName)
		}
	}
	return nil
}

func (c *k8sServiceCatalogClient) setSecret(organizationID string, secretName string, secrets map[string]string) error {
	log.Debugf("Setting dispatch secret %s", secretName)
	// We should probably update only on changes rather than just by default
	_, err := c.secretsClient.UpdateSecret(
		context.Background(),
		organizationID,
		&dispatchv1.Secret{
			Name:    &secretName,
			Secrets: secrets,
		},
	)
	if err != nil {
		log.Debugf("failed to update secrets in secret store: %v", err)
		// If update failed, probably missing so create
		_, err := c.secretsClient.CreateSecret(
			context.Background(),
			organizationID,
			&dispatchv1.Secret{
				Name:    &secretName,
				Secrets: secrets,
			},
		)
		if err != nil {
			return errors.Wrapf(err, "failed to set secrets in secret store")
		}

	}
	return nil
}

func buildEnv(input map[string]string) []corev1.EnvVar {
	var vars []corev1.EnvVar
	for key, val := range input {
		envVar := corev1.EnvVar{
			Name:  strings.Replace(strings.ToUpper(key), "-", "_", -1),
			Value: val,
		}
		vars = append(vars, envVar)
	}
	return vars
}

func buildArgs(input map[string]string) []string {
	var args []string
	for key, val := range input {
		args = append(args, fmt.Sprintf("--%s=%s", key, val))
	}
	return args
}
