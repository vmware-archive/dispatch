///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package kubeless

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/client/clientset/versioned/typed/kubeless/v1beta1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedExtensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	jsonContentType = "application/json"

	defaultCreateTimeout = 60 // seconds
)

// Config contains the Kubeless configuration
type Config struct {
	ImageRegistry   string
	RegistryAuth    string
	K8sConfig       string
	FuncNamespace   string
	CreateTimeout   *int
	ImagePullSecret string
}

type kubelessDriver struct {
	imageBuilder functions.ImageBuilder
	docker       *docker.Client

	deployments typedExtensionsv1beta1.DeploymentInterface
	functions   kubelessv1beta1.FunctionInterface
	fnNs        string

	createTimeout   int
	imagePullSecret string
}

type systemError struct {
	Err error `json:"err"`
}

func (err *systemError) Error() string {
	return err.Err.Error()
}

func (err *systemError) AsSystemErrorObject() interface{} {
	return err
}

func (err *systemError) StackTrace() errors.StackTrace {
	if e, ok := err.Err.(functions.StackTracer); ok {
		return e.StackTrace()
	}

	return nil
}

// New creates a new Kubeless driver
func New(config *Config) (functions.FaaSDriver, error) {
	dc, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get docker client")
	}

	k8sConf, err := kubeClientConfig(config.K8sConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	k8sClient := kubernetes.NewForConfigOrDie(k8sConf)
	kubelessCli := versioned.NewForConfigOrDie(k8sConf)

	fnNs := config.FuncNamespace
	if fnNs == "" {
		fnNs = "default"
	}

	d := &kubelessDriver{
		imageBuilder:  functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, dc),
		docker:        dc,
		deployments:   k8sClient.ExtensionsV1beta1().Deployments(fnNs),
		functions:     kubelessCli.KubelessV1beta1().Functions(fnNs),
		fnNs:          fnNs,
		createTimeout: defaultCreateTimeout,
	}
	if config.CreateTimeout != nil {
		d.createTimeout = *config.CreateTimeout
	}
	if config.ImagePullSecret != "" {
		d.imagePullSecret = config.ImagePullSecret
	}

	return d, nil
}

func kubeClientConfig(kubeConfPath string) (*rest.Config, error) {
	if kubeConfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfPath)
	}
	return rest.InClusterConfig()
}

func getID(id string) string {
	return fmt.Sprintf("kbls-%s", id)
}

func (d *kubelessDriver) Create(ctx context.Context, f *functions.Function, exec *functions.Exec) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	image, err := d.imageBuilder.BuildImage(ctx, "kubeless", f.FaasID, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
	}

	kf := v1beta1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: getID(f.FaasID),
		},
		Spec: v1beta1.FunctionSpec{
			Deployment: extensionsv1beta1.Deployment{
				Spec: extensionsv1beta1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Image: image},
							},
						},
					},
				},
			},
		},
	}
	_, err = d.functions.Create(&kf)
	if err != nil {
		return err
	}

	// make sure the function has started
	return utils.Backoff(time.Duration(d.createTimeout)*time.Second, func() error {
		deployment, err := d.deployments.Get(getID(f.FaasID), metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to read function deployment status: '%s'", f.Name)
		}

		if deployment.Status.AvailableReplicas > 0 {
			return nil
		}

		return errors.Errorf("function deployment not available: '%s'", f.Name)
	})
}

func (d *kubelessDriver) Delete(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()
	err := d.functions.Delete(getID(f.FaasID), &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (d *kubelessDriver) doHTTPReq(faasID string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", getID(faasID), d.fnNs), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Unable to create request %v", err)
	}
	req.Header.Add("Content-Type", jsonContentType)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: received error code %d: %s", resp.StatusCode, resp.Status)
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *kubelessDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		bytesIn, _ := json.Marshal(functions.Message{Context: ctx, Payload: in})
		res, err := d.doHTTPReq(e.FaasID, bytesIn)
		if err != nil {
			return nil, err
		}
		var out functions.Message
		if err := json.Unmarshal(res, &out); err != nil {
			return nil, &systemError{errors.Errorf("cannot JSON-parse result from OpenFaaS: %s %s", err, string(res))}
		}
		ctx.AddLogs(out.Context.Logs())
		ctx.SetError(out.Context.GetError())
		return out.Payload, nil
	}
}
