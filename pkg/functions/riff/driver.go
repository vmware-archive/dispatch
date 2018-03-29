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

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/riff/internal"
	"github.com/vmware/dispatch/pkg/trace"
)

const consumerGroupID = "dispatch-riff-driver"
const correlationIDHeader = "correlationId" // header propagated by riff function-sidecar

// Config contains the riff configuration
type Config struct {
	KafkaBrokers  []string
	ImageRegistry string
	RegistryAuth  string
	K8sConfig     string
	FuncNamespace string
	TemplateDir   string
}

type riffDriver struct {
	requester *internal.Requester

	imageRegistry string
	registryAuth  string

	imageBuilder functions.ImageBuilder
	docker       *docker.Client

	riffTalk *internal.RiffTalk
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

	requester, err := internal.NewRequester(correlationIDHeader, consumerGroupID, config.KafkaBrokers)
	if err != nil {
		return nil, err
	}

	d := &riffDriver{
		requester:     requester,
		imageRegistry: config.ImageRegistry,
		registryAuth:  config.RegistryAuth,
		docker:        dc,
		imageBuilder:  functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, config.TemplateDir, dc),
		riffTalk:      internal.NewRiffTalk(config.K8sConfig, config.FuncNamespace),
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
