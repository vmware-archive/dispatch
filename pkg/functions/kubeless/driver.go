///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package kubeless

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
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
	services    typedv1.ServiceInterface
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

// New creates a new OpenFaaS driver
func New(config *Config) (functions.FaaSDriver, error) {
	defer trace.Trace("")()
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
		services:      k8sClient.CoreV1().Services(fnNs),
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

func (d *kubelessDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Trace("kubeless.Create." + f.ID)()

	image, err := d.imageBuilder.BuildImage("kubeless", f.FaasID, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
	}

	kf := v1beta1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.Name,
			Labels: map[string]string{
				"ID": f.FaasID,
			},
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
		defer trace.Trace("")()

		deployment, err := d.deployments.Get(f.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to read function deployment status: '%s'", f.Name)
		}

		if deployment.Status.AvailableReplicas > 0 {
			return nil
		}

		return errors.Errorf("function deployment not available: '%s'", f.Name)
	})
}

func (d *kubelessDriver) Delete(f *functions.Function) error {
	defer trace.Trace("kubeless.Delete." + f.ID)()
	err := d.functions.Delete(f.Name, &metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (d *kubelessDriver) getFuncName(id string) (string, error) {
	list, err := d.functions.List(metav1.ListOptions{LabelSelector: fmt.Sprintf("ID=%s", id)})
	if err != nil {
		return "", err
	}
	if len(list.Items) != 1 {
		return "", fmt.Errorf("Unexpected amount of functions found %v", list.Items)
	}
	return list.Items[0].Name, nil
}

func (d *kubelessDriver) getHTTPReq(funcName, eventID, eventNamespace string, body []byte) (*http.Request, error) {
	svc, err := d.services.Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to find the service for function %s", funcName)
	}
	funcPort := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", funcName, d.fnNs, funcPort), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Unable to create request %v", err)
	}
	req.Header.Add("Content-Type", jsonContentType)
	return req, nil
}

func (d *kubelessDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		defer trace.Trace("kubeless.run." + e.FunctionID)()

		req := &http.Request{}
		name, err := d.getFuncName(e.FaasID)
		if err != nil {
			return nil, err
		}
		bytesIn, _ := json.Marshal(functions.Message{Context: ctx, Payload: in})
		req, err = d.getHTTPReq(name, e.RunID, "dispatch.vmware.github.io", bytesIn)
		if err != nil {
			return nil, err
		}
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
		var out functions.Message
		if err := json.Unmarshal(res, &out); err != nil {
			return nil, &systemError{errors.Errorf("cannot JSON-parse result from OpenFaaS: %s %s", err, string(res))}
		}
		ctx.AddLogs(out.Context.Logs())
		ctx.SetError(out.Context.GetError())
		return out.Payload, nil
	}
}
