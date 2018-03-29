///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package riff

import (
	"encoding/json"
	"net/http"

	"github.com/bsm/sarama-cluster"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/projectriff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	riffcs "github.com/projectriff/kubernetes-crds/pkg/client/clientset/versioned"
	riffv1 "github.com/projectriff/kubernetes-crds/pkg/client/clientset/versioned/typed/projectriff/v1"
	"github.com/projectriff/riff/message-transport/pkg/transport/kafka"
	log "github.com/sirupsen/logrus"
	kapi "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware/dispatch/pkg/config"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

const consumerGroupID = "dispatch-riff-driver"
const correlationIDHeader = "correlationId" // header propagated by riff function-sidecar

// Config contains the riff configuration
type Config struct {
	KafkaBrokers        []string
	ImageRegistry       string
	RegistryAuth        string
	K8sConfig           string
	FuncNamespace       string
	FuncDefaultLimits   *config.FunctionResources
	FuncDefaultRequests *config.FunctionResources
	TemplateDir         string
}

type riffDriver struct {
	requester *requester

	imageRegistry string
	registryAuth  string

	imageBuilder functions.ImageBuilder
	httpClient   *http.Client
	docker       *docker.Client

	topics    riffv1.TopicInterface
	functions riffv1.FunctionInterface

	funcDefaultResourceReq kapi.ResourceRequirements
}

func (d *riffDriver) Close() error {
	return d.requester.Close()
}

// New creates a new riff driver
func New(config *Config) (functions.FaaSDriver, error) {
	defer trace.Trace("")()
	dc, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get docker client")
	}
	riffClient := newRiffClient(config.K8sConfig)

	producer, err := kafka.NewProducer(config.KafkaBrokers)
	if err != nil {
		return nil, errors.Wrap(err, "could not get kafka producer")
	}

	consumer, err := kafka.NewConsumer(config.KafkaBrokers, consumerGroupID, []string{"replies"}, cluster.NewConfig())
	if err != nil {
		return nil, errors.Wrap(err, "could not get kafka consumer")
	}

	funcDefaultResourceReq := kapi.ResourceRequirements{}
	if config.FuncDefaultLimits != nil {
		funcDefaultResourceReq.Limits = kapi.ResourceList{
			kapi.ResourceCPU:    resource.MustParse(config.FuncDefaultLimits.CPU),
			kapi.ResourceMemory: resource.MustParse(config.FuncDefaultLimits.Memory)}
	}

	if config.FuncDefaultRequests != nil {
		funcDefaultResourceReq.Requests = kapi.ResourceList{
			kapi.ResourceCPU:    resource.MustParse(config.FuncDefaultRequests.CPU),
			kapi.ResourceMemory: resource.MustParse(config.FuncDefaultRequests.Memory)}
	}

	d := &riffDriver{
		requester:              newRequester(correlationIDHeader, producer, consumer),
		imageRegistry:          config.ImageRegistry,
		registryAuth:           config.RegistryAuth,
		httpClient:             http.DefaultClient,
		docker:                 dc,
		imageBuilder:           functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, config.TemplateDir, dc),
		topics:                 riffClient.ProjectriffV1().Topics(config.FuncNamespace),
		functions:              riffClient.ProjectriffV1().Functions(config.FuncNamespace),
		funcDefaultResourceReq: funcDefaultResourceReq,
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

func (d *riffDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Tracef("riff.Create.%s", f.ID)()

	image, err := d.imageBuilder.BuildImage("riff", f.ID, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
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
				Image:     image,
				Resources: d.funcDefaultResourceReq,
			},
		},
	}

	if _, err := d.topics.Create(topic); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "error creating topic '%s'", fnName)
		}
	}

	if _, err := d.functions.Create(function); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "error creating function '%s'", fnName)
		}
	}

	return nil
}

func (d *riffDriver) Delete(f *functions.Function) error {
	defer trace.Tracef("riff.Delete.%s", f.ID)()

	fnName := fnID(f.ID)

	if err := d.functions.Delete(fnName, nil); err != nil {
		if !kerrors.IsNotFound(err) {
			return errors.Wrapf(err, "error deleting function '%s'", fnName)
		}
	}

	if err := d.topics.Delete(fnName, nil); err != nil {
		if !kerrors.IsNotFound(err) {
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
		defer trace.Tracef("riff.run.%s", e.FunctionID)()

		bytesIn, _ := json.Marshal(ctxAndPld{Context: ctx, Payload: in})
		topic := fnID(e.FunctionID)

		log.Debugf("Posting to topic '%s': '%s'", topic, string(bytesIn))

		resBytes, err := d.requester.Request(topic, e.RunID, bytesIn)
		if err != nil {
			return nil, errors.Wrapf(err, "riff: error invoking function: '%s', runID: '%s'", e.FunctionID, e.RunID)
		}

		var out ctxAndPld
		if err := json.Unmarshal(resBytes, &out); err != nil {
			return nil, errors.Errorf("cannot JSON-parse result from riff: %s %s", err, string(resBytes))
		}
		ctx.AddLogs(out.Context.Logs())
		return out.Payload, nil

	}
}

func fnID(id string) string {
	return "fn-" + id
}
