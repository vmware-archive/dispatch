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

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/client/clientset/versioned/typed/kubeless/v1beta1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
	kapi "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	K8sConfig           string
	FuncNamespace       string
	CreateTimeout       *int
	ImagePullSecret     string
	FuncDefaultLimits   *functions.FunctionResources
	FuncDefaultRequests *functions.FunctionResources
}

type kubelessDriver struct {
	deployments typedExtensionsv1beta1.DeploymentInterface
	functions   kubelessv1beta1.FunctionInterface
	fnNs        string

	createTimeout   int
	imagePullSecret string

	funcDefaultLimits   *functions.FunctionResources
	funcDefaultRequests *functions.FunctionResources
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

	funcDefaultLimits := &functions.FunctionResources{}
	if config.FuncDefaultLimits != nil {
		funcDefaultLimits.CPU = config.FuncDefaultLimits.CPU
		funcDefaultLimits.Memory = config.FuncDefaultLimits.Memory
	}

	funcDefaultRequests := &functions.FunctionResources{}
	if config.FuncDefaultRequests != nil {
		funcDefaultRequests.CPU = config.FuncDefaultRequests.CPU
		funcDefaultRequests.Memory = config.FuncDefaultRequests.Memory
	}

	d := &kubelessDriver{
		deployments:         k8sClient.ExtensionsV1beta1().Deployments(fnNs),
		functions:           kubelessCli.KubelessV1beta1().Functions(fnNs),
		fnNs:                fnNs,
		createTimeout:       defaultCreateTimeout,
		funcDefaultLimits:   funcDefaultLimits,
		funcDefaultRequests: funcDefaultRequests,
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

func (d *kubelessDriver) Create(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	funcLimits := &functions.FunctionResources{
		CPU:    d.funcDefaultLimits.CPU,
		Memory: d.funcDefaultLimits.Memory,
	}
	if f.ResourceLimits.CPU != "" {
		funcLimits.CPU = f.ResourceLimits.CPU
	}
	if f.ResourceLimits.Memory != "" {
		funcLimits.Memory = f.ResourceLimits.Memory
	}

	funcRequests := &functions.FunctionResources{
		CPU:    d.funcDefaultRequests.CPU,
		Memory: d.funcDefaultRequests.Memory,
	}
	if f.ResourceRequests.CPU != "" {
		funcRequests.CPU = f.ResourceRequests.CPU
	}
	if f.ResourceRequests.Memory != "" {
		funcRequests.Memory = f.ResourceRequests.Memory
	}

	resourceRequirements := kapi.ResourceRequirements{
		Limits:   kapi.ResourceList{},
		Requests: kapi.ResourceList{},
	}
	if funcLimits.CPU != "" {
		qty, err := resource.ParseQuantity(funcLimits.CPU)
		if err != nil {
			return errors.Wrapf(err, "error parsing cpu limit '%s'", funcLimits.CPU)
		}
		resourceRequirements.Limits[kapi.ResourceCPU] = qty
	}
	if funcLimits.Memory != "" {
		qty, err := resource.ParseQuantity(funcLimits.Memory)
		if err != nil {
			return errors.Wrapf(err, "error parsing memory limit '%s'", funcLimits.Memory)
		}
		resourceRequirements.Limits[kapi.ResourceMemory] = qty
	}
	if funcRequests.CPU != "" {
		qty, err := resource.ParseQuantity(funcRequests.CPU)
		if err != nil {
			return errors.Wrapf(err, "error parsing cpu request '%s'", funcRequests.CPU)
		}
		resourceRequirements.Requests[kapi.ResourceCPU] = qty
	}
	if funcRequests.Memory != "" {
		qty, err := resource.ParseQuantity(funcRequests.Memory)
		if err != nil {
			return errors.Wrapf(err, "error parsing memory request '%s'", funcRequests.Memory)
		}
		resourceRequirements.Requests[kapi.ResourceMemory] = qty
	}

	kf := v1beta1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: getID(f.FaasID),
		},
		Spec: v1beta1.FunctionSpec{
			Deployment: extensionsv1beta1.Deployment{
				Spec: extensionsv1beta1.DeploymentSpec{
					Template: kapi.PodTemplateSpec{
						Spec: kapi.PodSpec{
							Containers: []kapi.Container{
								{
									Image:     f.FunctionImageURL,
									Resources: resourceRequirements,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := d.functions.Create(&kf)
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
