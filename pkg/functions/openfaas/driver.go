///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package openfaas

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

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
}

type imgResult struct {
	image string
	err   error
}

type imgRequest struct {
	name string
	exec *functions.Exec

	result chan *imgResult
}

func newFuncImgRequest(name string, exec *functions.Exec) *imgRequest {
	return &imgRequest{
		name:   name,
		exec:   exec,
		result: make(chan *imgResult),
	}
}

type ofDriver struct {
	gateway       string
	imageRegistry string
	registryAuth  string

	httpClient *http.Client
	docker     *docker.Client

	requests chan *imgRequest
}

func New(config *Config) (functions.FaaSDriver, error) {
	defer trace.Trace("")()
	dc, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get docker client")
	}
	d := &ofDriver{
		gateway:       strings.TrimRight(config.Gateway, "/"),
		imageRegistry: config.ImageRegistry,
		registryAuth:  config.RegistryAuth,
		httpClient:    http.DefaultClient,
		docker:        dc,
		requests:      make(chan *imgRequest),
	}
	go d.processRequests()

	return d, nil
}

func (d *ofDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Trace("openfaas.Create." + f.Name)()

	imgReq := newFuncImgRequest(f.Name, exec)
	d.requests <- imgReq
	result := <-imgReq.result

	if result.err != nil {
		return errors.Wrapf(result.err, "Error building image for function '%s'", f.Name)
	}

	if err := d.Delete(f); err != nil {
		return errors.Wrapf(err, "Failed to cleanup before deploying function '%s'", f.Name)
	}

	req := requests.CreateFunctionRequest{
		Image:       result.image,
		Network:     "func_functions",
		Service:     getID(f.ID),
		EnvVars:     map[string]string{},
		Constraints: []string{},
	}

	reqBytes, _ := json.Marshal(&req)
	res, err := d.httpClient.Post(d.gateway+"/system/functions", jsonContentType, bytes.NewReader(reqBytes))
	if err != nil {
		return errors.Wrapf(err, "Error deploying function '%s'", f.Name)
	}
	defer res.Body.Close()

	log.Debugf("openfaas.Create.%s: status code: %v", f.Name, res.StatusCode)
	switch res.StatusCode {
	case 200, 201, 202:
		return nil

	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return errors.Errorf("Server returned unexpected status: %v, %s", res.StatusCode, string(bytesOut))
		}
		return errors.Wrapf(err, "Error performing POST request, status: %v", res.StatusCode)
	}
}

func (d *ofDriver) Delete(f *functions.Function) error {
	defer trace.Trace("openfaas.Delete." + f.Name)()

	reqBytes, _ := json.Marshal(&requests.DeleteFunctionRequest{FunctionName: getID(f.ID)})
	req, _ := http.NewRequest("DELETE", d.gateway+"/system/functions", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", jsonContentType)

	res, err := d.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Error removing existing function: %s, gateway=%s, functionName=%s\n", err.Error(), d.gateway, f.Name)
	}
	defer res.Body.Close()

	log.Debugf("openfaas.Delete.%s: status code: %v", f.Name, res.StatusCode)
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

type ctxAndIn struct {
	Context functions.Context `json:"context"`
	Input   interface{}       `json:"input"`
}

const xStderrHeader = "X-Stderr"

func (d *ofDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		defer trace.Trace("openfaas.run." + e.Name)()

		bytesIn, _ := json.Marshal(ctxAndIn{Context: ctx, Input: in})
		res, err := d.httpClient.Post(d.gateway+"/function/"+getID(e.ID), jsonContentType, bytes.NewReader(bytesIn))
		if err != nil {
			return nil, errors.Errorf("cannot connect to OpenFaaS on URL: %s", d.gateway)
		}
		defer res.Body.Close()

		log.Debugf("openfaas.run.%s: status code: %v", e.Name, res.StatusCode)
		switch res.StatusCode {
		case 200:
			ctx.ReadLogs(logsReader(res))
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, errors.Errorf("cannot read result from OpenFaaS on URL: %s %s", d.gateway, err)
			}
			var out interface{}
			if err := json.Unmarshal(resBytes, &out); err != nil {
				return nil, errors.Errorf("cannot JSON-parse result from OpenFaaS: %s %s", err, string(resBytes))
			}
			return out, nil

		default:
			bytesOut, err := ioutil.ReadAll(res.Body)
			if err == nil {
				return nil, errors.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
			}
			return nil, errors.Wrapf(err, "Error performing DELETE request, status: %v", res.StatusCode)
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
