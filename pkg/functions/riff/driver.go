///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package riff

import (
	"encoding/json"

	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/vmware/dispatch/lib/riff"
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
}

type riffDriver struct {
	requester *riff.Requester

	imageRegistry string
	registryAuth  string

	imageBuilder functions.ImageBuilder
	docker       *docker.Client

	riffTalk *riff.RiffTalk
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

	requester, err := riff.NewRequester(correlationIDHeader, consumerGroupID, config.KafkaBrokers)
	if err != nil {
		return nil, err
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
		requester:     requester,
		imageRegistry: config.ImageRegistry,
		registryAuth:  config.RegistryAuth,
		docker:        dc,
		imageBuilder:  functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, dc),
		riffTalk:      riff.NewRiffTalk(config.K8sConfig, config.FuncNamespace),
	}

	return d, nil
}

func (d *riffDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Tracef("riff.Create.%s", f.ID)()

	image, err := d.imageBuilder.BuildImage("riff", f.ID, exec)
	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
	}

	return d.riffTalk.Create(fnID(f.ID), image)
}

func (d *riffDriver) Delete(f *functions.Function) error {
	defer trace.Tracef("riff.Delete.%s", f.ID)()

	return d.riffTalk.Delete(fnID(f.ID))
}

func (d *riffDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		defer trace.Tracef("riff.run.%s", e.FunctionID)()

		bytesIn, _ := json.Marshal(functions.Message{Context: ctx, Payload: in})
		topic := fnID(e.FunctionID)

		log.Debugf("Posting to topic '%s': '%s'", topic, string(bytesIn))

		resBytes, err := d.requester.Request(topic, e.RunID, bytesIn)
		if err != nil {
			return nil, &systemError{errors.Wrapf(err, "riff: error invoking function: '%s', runID: '%s'", e.FunctionID, e.RunID)}
		}

		var out functions.Message
		if err := json.Unmarshal(resBytes, &out); err != nil {
			return nil, &systemError{errors.Errorf("cannot JSON-parse result from riff: %s %s", err, string(resBytes))}
		}
		ctx.AddLogs(out.Context.Logs())
		ctx.SetError(out.Context.GetError())
		return out.Payload, nil
	}
}

func fnID(id string) string {
	return "fn-" + id
}
