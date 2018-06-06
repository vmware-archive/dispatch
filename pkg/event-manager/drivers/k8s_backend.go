///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

// NO TEST

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/utils"

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

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	eventDriverLabel     = "event-driver"
	defaultDeployTimeout = 10 // seconds
)

type k8sBackend struct {
	clientset     *kubernetes.Clientset
	config        ConfigOpts
	secretsClient client.SecretsClient
	DeployTimeout int
}

// NewK8sBackend creates a new K8s backend driver
func NewK8sBackend(secretsClient client.SecretsClient, config ConfigOpts) (Backend, error) {

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
		return nil, ewrapper.Wrap(err, "Error getting kubernetes config")
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, ewrapper.Wrap(err, "Error getting kubernetes clientset")
	}

	return &k8sBackend{
		clientset:     clientset,
		config:        config,
		secretsClient: secretsClient,
		DeployTimeout: defaultDeployTimeout,
	}, nil
}

func getDriverFullName(driver *entities.Driver) string {
	return fmt.Sprintf("event-driver-%s-%s", driver.Type, driver.Name)
}

func (k *k8sBackend) makeDeploymentSpec(secrets map[string]string, driver *entities.Driver) (*v1beta1.Deployment, error) {
	fullname := getDriverFullName(driver)

	// holds all inputs dedicated for event driver
	inputMap := map[string]string{}

	for key, val := range driver.Config {
		inputMap[key] = val
	}

	driverArgs := []string{driver.Type}

	driverImage := k.config.DriverImage
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
			Replicas: swag.Int32(1),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: fullname,
					Labels: map[string]string{
						"app": eventDriverLabel,
					},
				},
				Spec: corev1.PodSpec{
					Volumes:        nil,
					InitContainers: nil,
					Containers: []corev1.Container{
						{
							Name:            "driver",
							Image:           driverImage,
							Args:            driverArgs,
							Env:             buildEnv(secrets),
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
						{
							Name:            "driver-sidecar",
							Image:           k.config.SidecarImage,
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

func (k *k8sBackend) Deploy(ctx context.Context, driver *entities.Driver) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	secrets, err := k.getSecrets(ctx, driver.Secrets)
	if err != nil {
		return ewrapper.Wrapf(err, "failed to retrieve secrets")
	}

	deploymentSpec, err := k.makeDeploymentSpec(secrets, driver)
	if err != nil {
		err = &errors.DriverError{
			Err: ewrapper.Wrapf(err, "k8s: error making a deployment"),
		}
		log.Errorln(err)
		return err
	}

	result, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.DriverNamespace).Create(deploymentSpec)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			err = &EventdriverErrorDeploymentAlreadyExists{
				Err: ewrapper.Wrapf(err, "k8s: error creating a deployment"),
			}
		}
		log.Errorln(err)
		return err
	}

	if output, err := json.MarshalIndent(result, "", "  "); err == nil {
		log.Debugf("k8s: creating deployment\n%s\n", output)
	} else {
		log.Debugf("k8s: json marshal error")
	}

	return utils.Backoff(time.Duration(k.DeployTimeout)*time.Second, func() error {
		fullname := getDriverFullName(driver)
		deployment, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.DriverNamespace).Get(fullname, metav1.GetOptions{})

		if err != nil {
			if k8serrors.IsNotFound(err) {
				return &EventdriverErrorDeploymentNotFound{
					Err: ewrapper.Wrapf(err, "k8s: deployment=%s not found", fullname),
				}
			}
			return &EventdriverErrorUnknown{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s unexpected error", fullname),
			}
		}

		if deployment.Status.AvailableReplicas > 0 {
			return nil
		}

		return &EventdriverErrorDeploymentNotAvaialble{
			Err: ewrapper.Errorf("k8s: deployment=%s not available, pulling status", fullname),
		}
	})
}

func isEventDriver(deployment *v1beta1.Deployment) bool {
	if deployment != nil {
		val, ok := deployment.Labels["app"]
		if ok && val == eventDriverLabel {
			return true
		}
	}
	return false
}

func (k *k8sBackend) Delete(ctx context.Context, driver *entities.Driver) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	fullname := getDriverFullName(driver)

	deployment, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.DriverNamespace).Get(fullname, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			err = &EventdriverErrorDeploymentNotFound{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s not found", fullname),
			}
		} else {
			err = &EventdriverErrorUnknown{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s unexpected error", fullname),
			}
		}
		log.Errorln(err)
		return err
	}

	if !isEventDriver(deployment) {
		err = &errors.DriverError{
			Err: ewrapper.Wrapf(err, "k8s: deployment=%s: deleting a NON-event-driver deployment", fullname),
		}
		return err
	}

	foreground := metav1.DeletePropagationForeground
	if err := k.clientset.ExtensionsV1beta1().Deployments(k.config.DriverNamespace).Delete(fullname,
		&metav1.DeleteOptions{
			PropagationPolicy: &foreground,
		}); err != nil {

		if k8serrors.IsNotFound(err) {
			err = &errors.ObjectNotFoundError{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s not found", fullname),
			}
		} else {
			err = &errors.DriverError{
				Err: ewrapper.Wrapf(err, "k8s: deployment=%s unexpected error", fullname),
			}
		}
		log.Errorln(err)
		return err
	}
	return nil
}

// TODO: find better way of getting pod failure reason
func getReasonForUnavailablePods(pods *corev1.PodList, fullname string) error {
	for _, pod := range pods.Items {
		if !strings.Contains(pod.Name, fullname) {
			continue
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name != "driver" {
				continue
			}

			if containerStatus.State.Waiting != nil {
				waiting := containerStatus.State.Waiting
				return &EventdriverErrorDeploymentNotAvaialble{
					Err: ewrapper.Errorf("k8s: deployment `%s` is `waiting`: %s, %s",
						fullname, waiting.Reason, waiting.Message),
				}
			}
			if containerStatus.State.Terminated != nil {
				terminated := containerStatus.State.Terminated
				return &EventdriverErrorDeploymentNotAvaialble{
					Err: ewrapper.Errorf("k8s: deployment `%s` is `terminated`: %s, %s",
						fullname, terminated.Reason, terminated.Message),
				}
			}

		}
	}

	return nil
}

func (k *k8sBackend) Update(ctx context.Context, driver *entities.Driver) error {
	secrets, err := k.getSecrets(ctx, driver.Secrets)
	if err != nil {
		return ewrapper.Wrapf(err, "failed to retrieve secrets")
	}

	deploymentSpec, err := k.makeDeploymentSpec(secrets, driver)
	if err != nil {
		err = &errors.DriverError{
			Err: ewrapper.Wrapf(err, "k8s: error making a deployment"),
		}
		log.Errorln(err)
		return err
	}
	fullname := getDriverFullName(driver)

	if driver.GetStatus() == entitystore.StatusUPDATING {
		// In UPDATING status, do backend deployment Update()
		result, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.DriverNamespace).Update(deploymentSpec)
		if err != nil {
			err = &errors.DriverError{
				Err: ewrapper.Wrapf(err, "k8s: error updating a deployment"),
			}
			log.Errorln(err)
			return err
		}

		if output, err := json.MarshalIndent(result, "", "  "); err == nil {
			log.Debugf("k8s: updating deployment\n%s\n", output)
		} else {
			log.Debugf("k8s: json marshal error")
		}
	} else {
		// check avaiable replicasets
		deployment, err := k.clientset.ExtensionsV1beta1().Deployments(k.config.DriverNamespace).Get(fullname, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				err = &EventdriverErrorDeploymentNotFound{
					Err: ewrapper.Wrapf(err, "k8s: deployment=%s not found", fullname),
				}
			} else {
				err = &EventdriverErrorUnknown{
					Err: ewrapper.Wrapf(err, "k8s: deployment=%s unexpected error", fullname),
				}
			}
			log.Errorln(err)
			return err
		}
		if deployment.Status.AvailableReplicas == 0 {
			// Try to get reason from pod
			pods, err := k.clientset.CoreV1().Pods(k.config.DriverNamespace).List(metav1.ListOptions{
				LabelSelector: "app=" + eventDriverLabel,
			})
			if err != nil {
				log.Errorln(err)
			} else {
				if err := getReasonForUnavailablePods(pods, fullname); err != nil {
					return err
				}
			}
			err = &EventdriverErrorDeploymentNotAvaialble{
				Err: ewrapper.Errorf("k8s: deployment=%s not available", fullname),
			}
			log.Errorln(err)
			return err
		}
	}

	log.Debugf("k8s: deployment=%s updated", fullname)

	return nil
}

func (k *k8sBackend) getSecrets(ctx context.Context, secretNames []string) (map[string]string, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	secrets := make(map[string]string)
	for _, name := range secretNames {
		resp, err := k.secretsClient.GetSecret(ctx, k.config.OrgID, name)
		if err != nil {
			return secrets, ewrapper.Wrapf(err, "failed to get secrets from secret store")
		}
		for key, value := range resp.Secrets {
			secrets[key] = value
		}
	}
	return secrets, nil
}

func (k *k8sBackend) buildSidecarEnv(d *entities.Driver) []corev1.EnvVar {
	vars := []corev1.EnvVar{
		{
			Name:  "DISPATCH_KAFKA_BROKERS",
			Value: strings.Join(k.config.KafkaBrokers, ","),
		},
		{
			Name:  "DISPATCH_RABBITMQ_URL",
			Value: k.config.RabbitMQURL,
		},
		{
			Name:  "DISPATCH_TENANT",
			Value: k.config.OrgID,
		},
		{
			Name:  "DISPATCH_TRANSPORT",
			Value: k.config.TransportType,
		},
		{
			Name:  "DISPATCH_TRACER",
			Value: k.config.Tracer,
		},
		{
			Name:  "DISPATCH_DRIVER_TYPE",
			Value: d.Type,
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
		if val == "" {
			args = append(args, fmt.Sprintf("--%s", key))
		} else {
			args = append(args, fmt.Sprintf("--%s=%s", key, val))
		}

	}
	return args
}
