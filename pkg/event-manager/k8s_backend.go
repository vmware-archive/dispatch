///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

// NO TEST

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware/dispatch/pkg/errors"
	secretsclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
)

// DriverBackend defines the event driver backend interface
type DriverBackend interface {
	Deploy(*Driver) error
	Update(*Driver) error
	Delete(*Driver) error
}

// K8sBackendConfig specify options for Kubernetes driver backend
type K8sBackendConfig struct {
	Replicas             int32
	EnableReadinessProbe bool
	Namespace            string
}

type k8sBackend struct {
	clientset     *kubernetes.Clientset
	config        K8sBackendConfig
	secretsClient *secretsclient.SecretStore
}

// NewK8sBackend creates a new K8s backend driver
func NewK8sBackend() (DriverBackend, error) {

	var err error
	var config *rest.Config
	if EventManagerFlags.K8sConfig == "" {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
	} else {
		// create from a configuration
		config, err = clientcmd.BuildConfigFromFlags("", EventManagerFlags.K8sConfig)
	}
	if err != nil {
		return nil, ewrapper.Wrap(err, "Error getting kubernetes config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, ewrapper.Wrap(err, "Error getting kubernetes clientset")
	}

	return &k8sBackend{
		clientset: clientset,
		config: K8sBackendConfig{
			EnableReadinessProbe: false,
			Replicas:             1,
			Namespace:            EventManagerFlags.K8sNamespace,
		},
		secretsClient: SecretStoreClient(),
	}, nil
}

// SecretStoreClient returns a client to the secret store
func SecretStoreClient() *secretsclient.SecretStore {
	transport := apiclient.New(EventManagerFlags.SecretStore, secretsclient.DefaultBasePath, []string{"http"})
	return secretsclient.New(transport, strfmt.Default)
}

func getDriverFullName(driver *Driver) string {
	return fmt.Sprintf("event-driver-%s-%s", driver.Type, driver.Name)
}

func (k *k8sBackend) makeDeploymentSpec(driver *Driver) (*v1beta1.Deployment, error) {
	fullname := getDriverFullName(driver)

	secrets, err := k.getSecrets(driver.Secrets)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "failed to retrieve secrets")
	}

	// holds all inputs dedicated for event driver
	inputMap := map[string]string{}

	for key, val := range driver.Config {
		inputMap[key] = val
	}

	// secrets are more important than config
	for key, val := range secrets {
		inputMap[key] = val
	}

	driverArgs := []string{driver.Type}

	driverImage := EventManagerFlags.EventDriverImage
	if _, ok := builtInDrivers[driver.Type]; !ok {
		driverImage = driver.Image
		driverArgs = buildArgs(inputMap)
	} else {
		driverArgs = append(driverArgs, buildArgs(inputMap)...)
	}

	deploymentSpec := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fullname,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: swag.Int32(k.config.Replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: fullname,
					Labels: map[string]string{
						"app": "event-driver",
					},
				},
				Spec: corev1.PodSpec{
					Volumes:        nil,
					InitContainers: nil,
					Containers: []corev1.Container{
						{
							Name:            fullname,
							Image:           driverImage,
							Args:            driverArgs,
							Env:             buildEnv(inputMap),
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
						{
							Name:            fullname + "-sidecar",
							Image:           EventManagerFlags.EventSidecarImage,
							Env:             k.buildSidecarEnv(driver),
							Resources:       corev1.ResourceRequirements{},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}
	return deploymentSpec, nil
}

func (k *k8sBackend) Deploy(driver *Driver) error {

	deploymentSpec, err := k.makeDeploymentSpec(driver)
	if err != nil {
		err = &errors.DriverError{
			Err: ewrapper.Wrapf(err, "k8s: error making a deployment"),
		}
		log.Errorln(err)
		return err
	}

	result, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.Namespace).Create(deploymentSpec)
	if err != nil {
		err = &errors.DriverError{
			Err: ewrapper.Wrapf(err, "k8s: error creating a deployment"),
		}
		log.Errorln(err)
		return err
	}

	if output, err := json.MarshalIndent(result, "", "  "); err == nil {
		log.Debugf("k8s: creating deployment\n%s\n", output)
	} else {
		log.Debugf("k8s: json marshal error")
	}

	log.Debugf("k8s: deployment=%s created", getDriverFullName(driver))
	return nil
}

func isEventDriver(deployment *v1beta1.Deployment) bool {
	if deployment != nil {
		val, ok := deployment.Labels["app"]
		if ok && val == "event-driver" {
			return true
		}
	}
	return false
}

func (k *k8sBackend) Delete(driver *Driver) error {

	fullname := getDriverFullName(driver)

	deployment, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.Namespace).Get(fullname, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			err = &errors.ObjectNotFoundError{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s not found", fullname),
			}
		} else {
			err = &errors.DriverError{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s unexpected error"),
			}
		}
		log.Errorln(err)
		return err
	}

	if !isEventDriver(deployment) {
		err = &errors.DriverError{
			Err: ewrapper.Wrapf(err, "k8s: deployment=%s: deleting a NON-event-driver deployment"),
		}
		return err
	}

	foreground := metav1.DeletePropagationForeground
	if err := k.clientset.ExtensionsV1beta1().Deployments(k.config.Namespace).Delete(fullname,
		&metav1.DeleteOptions{
			PropagationPolicy: &foreground,
		}); err != nil {

		if k8serrors.IsNotFound(err) {
			err = &errors.ObjectNotFoundError{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s not found", fullname),
			}
		} else {
			err = &errors.DriverError{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s unexpected error"),
			}
		}
		log.Errorln(err)
		return err
	}
	return nil
}

func (k *k8sBackend) Update(driver *Driver) error {
	// TODO:
	return fmt.Errorf("Update not implemented yet")
}

func (k *k8sBackend) getSecrets(secretNames []string) (map[string]string, error) {

	secrets := make(map[string]string)
	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", "cookie")
	for _, name := range secretNames {
		resp, err := k.secretsClient.Secret.GetSecret(&secret.GetSecretParams{
			SecretName: name,
			Context:    context.Background(),
		}, apiKeyAuth)
		if err != nil {
			return secrets, ewrapper.Wrapf(err, "failed to get secrets from secret store")
		}
		for key, value := range resp.Payload.Secrets {
			secrets[key] = value
		}
	}
	return secrets, nil
}

func (k *k8sBackend) buildSidecarEnv(d *Driver) []corev1.EnvVar {
	vars := []corev1.EnvVar{
		{
			Name:  "DISPATCH_RABBITMQ_URL",
			Value: EventManagerFlags.RabbitMQURL,
		},
		{
			Name:  "DISPATCH_TENANT",
			Value: EventManagerFlags.OrgID,
		},
		{
			Name:  "DISPATCH_TRACER_URL",
			Value: EventManagerFlags.TracerURL,
		},
	}
	if _, ok := builtInDrivers[d.Type]; !ok {
		// TODO handle custom drivers
	}
	return vars
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
