///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package noop

// The no-op driver is just a simple driver which essentially does nothing.
// It's useful for testing the function manager without requiring the FaaS.

// NO TESTS

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	jsonContentType = "application/json"
)

// Config for the no-op function driver
type Config struct {
	ImageRegistry string
	RegistryAuth  string
	TemplateDir   string
}

type noopDriver struct {
	imageBuilder functions.ImageBuilder
	docker       *docker.Client
}

// New is the constructor for the no-op function driver
func New(config *Config) (functions.FaaSDriver, error) {
	defer trace.Trace("")()
	dc, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get docker client")
	}

	d := &noopDriver{
		imageBuilder: functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, config.TemplateDir, dc),
		docker:       dc,
	}

	return d, nil
}

// Create creates a function [image] but no actual function
func (d *noopDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Trace("noop.Create." + f.ID)()

	_, err := d.imageBuilder.BuildImage("noop", f.ID, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
	}
	return nil
}

// Delete is a no-op
func (d *noopDriver) Delete(f *functions.Function) error {
	defer trace.Trace("noop.Delete." + f.ID)()
	return nil
}

type ctxAndIn struct {
	Context functions.Context `json:"context"`
	Input   interface{}       `json:"input"`
}

const xStderrHeader = "X-Stderr"

// GetRunnable returns a functions.Runnable
func (d *noopDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		defer trace.Trace("noop.run." + e.FunctionID)()
		return nil, nil
	}
}

func getID(id string) string {
	return fmt.Sprintf("noop-%s", id)
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
