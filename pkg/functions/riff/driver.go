///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package riff

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/vmware/dispatch/lib/riff"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

const consumerGroupID = "dispatch-riff-driver"
const correlationIDHeader = "correlationId" // header propagated by riff function-sidecar

// Config contains the riff configuration
type Config struct {
	KafkaBrokers        []string
	K8sConfig           string
	FuncNamespace       string
	FuncDefaultLimits   *functions.FunctionResources
	FuncDefaultRequests *functions.FunctionResources
}

type riffDriver struct {
	requester *riff.Requester
	riffTalk  *riff.RiffTalk
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
		requester: requester,
		riffTalk:  riff.NewRiffTalk(config.K8sConfig, config.FuncNamespace),
	}

	return d, nil
}

func (d *riffDriver) Create(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return d.riffTalk.Create(fnID(f.FaasID), f.FunctionImageURL)
}

func (d *riffDriver) Delete(ctx context.Context, f *functions.Function) error {
	return d.riffTalk.Delete(fnID(f.FaasID))
}

func (d *riffDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {

		bytesIn, _ := json.Marshal(functions.Message{Context: ctx, Payload: in})
		topic := fnID(e.FaasID)

		log.Infof("Posting to topic '%s': '%s'", topic, string(bytesIn))

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
