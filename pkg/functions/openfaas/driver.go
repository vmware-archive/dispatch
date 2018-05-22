///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package openfaas

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware/dispatch/pkg/config"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	jsonContentType = "application/json"

	defaultCreateTimeout = 60 // seconds
)

// Config contains the OpenFaaS configuration
type Config struct {
	Gateway             string
	ImageRegistry       string
	RegistryAuth        string
	K8sConfig           string
	FuncNamespace       string
	FuncDefaultLimits   *config.FunctionResources
	FuncDefaultRequests *config.FunctionResources
	CreateTimeout       *int
	ImagePullSecret     string
}

type ofDriver struct {
	gateway string

	imageBuilder functions.ImageBuilder
	httpClient   *http.Client
	docker       *docker.Client

	deployments v1beta1.DeploymentInterface

	createTimeout   int
	imagePullSecret string

	funcDefaultLimits   *requests.FunctionResources
	funcDefaultRequests *requests.FunctionResources
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
	dc, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get docker client")
	}

	k8sConf, err := kubeClientConfig(config.K8sConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	k8sClient := kubernetes.NewForConfigOrDie(k8sConf)

	fnNs := config.FuncNamespace
	if fnNs == "" {
		fnNs = "default"
	}

	var funcDefaultLimits *requests.FunctionResources
	if config.FuncDefaultLimits != nil {
		funcDefaultLimits = &requests.FunctionResources{
			CPU:    config.FuncDefaultLimits.CPU,
			Memory: config.FuncDefaultLimits.Memory,
		}
	}

	var funcDefaultRequests *requests.FunctionResources
	if config.FuncDefaultRequests != nil {
		funcDefaultRequests = &requests.FunctionResources{
			CPU:    config.FuncDefaultRequests.CPU,
			Memory: config.FuncDefaultRequests.Memory,
		}
	}

	d := &ofDriver{
		gateway:      strings.TrimRight(config.Gateway, "/"),
		httpClient:   http.DefaultClient,
		imageBuilder: functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, dc),
		docker:       dc,
		// Use AppsV1beta1 until we remove support for Kubernetes 1.7
		deployments:         k8sClient.AppsV1beta1().Deployments(fnNs),
		funcDefaultLimits:   funcDefaultLimits,
		funcDefaultRequests: funcDefaultRequests,
		createTimeout:       defaultCreateTimeout,
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

func (d *ofDriver) Create(ctx context.Context, f *functions.Function, exec *functions.Exec) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	image, err := d.imageBuilder.BuildImage(ctx, "openfaas", f.FaasID, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
	}

	req := requests.CreateFunctionRequest{
		Image:       image,
		Network:     "func_functions",
		Service:     getID(f.FaasID),
		EnvVars:     map[string]string{},
		Constraints: []string{},
		Limits:      d.funcDefaultLimits,
		Requests:    d.funcDefaultRequests,
	}
	if d.imagePullSecret != "" {
		req.Secrets = []string{d.imagePullSecret}
	}

	reqBytes, _ := json.Marshal(&req)
	res, err := d.httpClient.Post(d.gateway+"/system/functions", jsonContentType, bytes.NewReader(reqBytes))
	if err != nil {
		return errors.Wrapf(err, "Error deploying function '%s'", f.ID)
	}
	defer res.Body.Close()

	log.Debugf("openfaas.Create.%s: status code: %v", f.ID, res.StatusCode)
	switch res.StatusCode {
	case 200, 201, 202:
		// OK

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return errors.Errorf("Server returned unexpected status: %v, %s", res.StatusCode, string(bytesOut))
		}
		return errors.Wrapf(err, "Error performing POST request, status: %v", res.StatusCode)
	}

	// make sure the function has started
	return utils.Backoff(time.Duration(d.createTimeout)*time.Second, func() error {
		deployment, err := d.deployments.Get(getID(f.FaasID), v1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to read function deployment status: '%s'", getID(f.FaasID))
		}

		if deployment.Status.AvailableReplicas > 0 {
			return nil
		}

		return errors.Errorf("function deployment not available: '%s'", getID(f.FaasID))
	})
}

func (d *ofDriver) Delete(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	reqBytes, _ := json.Marshal(&requests.DeleteFunctionRequest{FunctionName: getID(f.FaasID)})
	req, _ := http.NewRequest("DELETE", d.gateway+"/system/functions", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", jsonContentType)

	res, err := d.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Error removing existing function: %s, gateway=%s, functionName=%s\n", err.Error(), d.gateway, f.FaasID)
	}
	defer res.Body.Close()

	log.Debugf("openfaas.Delete.%s: status code: %v", f.ID, res.StatusCode)
	switch res.StatusCode {
	case 200, 201, 202, 404, 500:
		return nil
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return errors.Errorf("Server returned unexpected status: %v, %s", res.StatusCode, string(bytesOut))
		}
		return errors.Wrapf(err, "Error performing DELETE request, status: %v", res.StatusCode)
	}
}

const xStderrHeader = "X-Stderr"

func (d *ofDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		bytesIn, _ := json.Marshal(functions.Message{Context: ctx, Payload: in})
		postURL := d.gateway + "/function/" + getID(e.FaasID)
		res, err := d.httpClient.Post(postURL, jsonContentType, bytes.NewReader(bytesIn))
		if err != nil {
			log.Errorf("Error when sending POST request to %s: %+v", postURL, err)
			return nil, &systemError{errors.Wrapf(err, "request to OpenFaaS on %s failed", d.gateway)}
		}
		defer res.Body.Close()

		log.Debugf("openfaas.run.%s: status code: %v", e.FunctionID, res.StatusCode)
		switch res.StatusCode {
		case 200:
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, &systemError{errors.Errorf("cannot read result from OpenFaaS on URL: %s %s", d.gateway, err)}
			}
			var out functions.Message
			if err := json.Unmarshal(resBytes, &out); err != nil {
				return nil, &systemError{errors.Errorf("cannot JSON-parse result from OpenFaaS: %s %s", err, string(resBytes))}
			}
			ctx.AddLogs(out.Context.Logs())
			ctx.SetError(out.Context.GetError())
			return out.Payload, nil

		default:
			bytesOut, err := ioutil.ReadAll(res.Body)
			if err == nil {
				return nil, &systemError{errors.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))}
			}
			return nil, &systemError{errors.Wrapf(err, "Error performing request, status: %v", res.StatusCode)}
		}
	}
}

func getID(id string) string {
	return fmt.Sprintf("of-%s", id)
}

func logsReader(res *http.Response) io.Reader {
	bs := base64Decode(res.Header.Get(xStderrHeader))
	return bytes.NewReader(bs)
}

func base64Decode(b64s string) []byte {
	b64dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64s))
	bs, _ := ioutil.ReadAll(b64dec)
	return bs
}
