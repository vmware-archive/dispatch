///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package riff

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/projectriff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	riffcs "github.com/projectriff/kubernetes-crds/pkg/client/clientset/versioned"
	riffv1 "github.com/projectriff/kubernetes-crds/pkg/client/clientset/versioned/typed/projectriff/v1"
	log "github.com/sirupsen/logrus"
	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	jsonContentType = "application/json"
)

type Config struct {
	Gateway       string
	ImageRegistry string
	RegistryAuth  string
	K8sConfig     string
	RiffNamespace string
}

type riffDriver struct {
	httpGateway string

	imageBuilder functions.ImageBuilder
	httpClient   *http.Client

	topics    riffv1.TopicInterface
	functions riffv1.FunctionInterface
}

func New(config *Config) (functions.FaaSDriver, error) {
	defer trace.Trace("")()
	builder, err := functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating docker image builder")
	}
	riffClient := newRiffClient(config.K8sConfig)

	d := &riffDriver{
		httpGateway:  strings.TrimRight(config.Gateway, "/"),
		httpClient:   http.DefaultClient,
		imageBuilder: builder,
		topics:       riffClient.ProjectriffV1().Topics(config.RiffNamespace),
		functions:    riffClient.ProjectriffV1().Functions(config.RiffNamespace),
	}

	return d, nil
}

func kubeClientConfig(kubeconfPath string) (*rest.Config, error) {
	if kubeconfPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfPath)
	}
	return rest.InClusterConfig()
}

func newRiffClient(kubeconfPath string) riffcs.Interface {
	config, err := kubeClientConfig(kubeconfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	return riffcs.NewForConfigOrDie(config)
}

type statusError interface {
	Status() metav1.Status
}

func (d *riffDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Tracef("riff.Create.%s", f.Name)()

	image, err := d.imageBuilder.BuildImage(f.Name, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.Name)
	}

	fnName := fnID(f.ID)

	topic := &v1.Topic{
		ObjectMeta: metav1.ObjectMeta{
			Name: fnName,
		},
	}
	function := &v1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: fnName,
		},
		Spec: v1.FunctionSpec{
			Protocol: "http",
			Input:    fnName,
			Container: kapi.Container{
				Image: image,
			},
		},
	}

	if _, err := d.topics.Create(topic); err != nil {
		statusErr, ok := err.(statusError)
		if !ok || statusErr.Status().Reason != "AlreadyExists" {
			return errors.Wrapf(err, "error creating topic '%s'", fnName)
		}
	}

	if _, err := d.functions.Create(function); err != nil {
		statusErr, ok := err.(statusError)
		if !ok || statusErr.Status().Reason != "AlreadyExists" {
			return errors.Wrapf(err, "error creating function '%s'", fnName)
		}
		if err := d.functions.Delete(fnName, nil); err != nil {
			return errors.Wrapf(err, "error deleting function '%s'", fnName)
		}
		if _, err := d.functions.Create(function); err != nil {
			return errors.Wrapf(err, "error updating function '%s'", fnName)
		}
	}

	return nil
}

func (d *riffDriver) Delete(f *functions.Function) error {
	defer trace.Tracef("riff.Delete.%s", f.Name)()

	fnName := fnID(f.ID)

	if err := d.functions.Delete(fnName, nil); err != nil {
		statusErr, ok := err.(statusError)
		if !ok || statusErr.Status().Reason != "NotFound" {
			return errors.Wrapf(err, "error deleting function '%s'", fnName)
		}
	}

	if err := d.topics.Delete(fnName, nil); err != nil {
		statusErr, ok := err.(statusError)
		if !ok || statusErr.Status().Reason != "NotFound" {
			return errors.Wrapf(err, "error deleting topic '%s'", fnName)
		}
	}

	return nil
}

type ctxAndPld struct {
	Context functions.Context `json:"context"`
	Payload interface{}       `json:"payload"`
}

func (d *riffDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		defer trace.Tracef("riff.run.%s", e.Name)()

		bytesIn, _ := json.Marshal(ctxAndPld{Context: ctx, Payload: in})
		req, _ := http.NewRequest("POST", d.httpGateway+"/requests/"+fnID(e.ID), bytes.NewReader(bytesIn))
		req.Header.Set("Content-Type", jsonContentType)
		req.Header.Set("Accept", jsonContentType)
		res, err := d.httpClient.Do(req)
		if err != nil {
			return nil, errors.Errorf("cannot connect to riff on URL: %s", d.httpGateway)
		}
		defer res.Body.Close()

		log.Debugf("riff.run.%s: status code: %v", e.Name, res.StatusCode)
		switch res.StatusCode {
		case 200:
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, errors.Errorf("cannot read result from riff on URL: %s %s", d.httpGateway, err)
			}
			var out ctxAndPld
			if err := json.Unmarshal(resBytes, &out); err != nil {
				return nil, errors.Errorf("cannot JSON-parse result from riff: %s %s", err, string(resBytes))
			}
			ctx.AddLogs(out.Context.Logs())
			return out.Payload, nil

		default:
			bytesOut, err := ioutil.ReadAll(res.Body)
			if err == nil {
				return nil, errors.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
			}
			return nil, errors.Wrapf(err, "Error performing request, status: %v", res.StatusCode)
		}
	}
}

func fnID(id string) string {
	return "fn-" + id
}
