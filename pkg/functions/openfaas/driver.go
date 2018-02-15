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

type ofDriver struct {
	gateway string

	imageBuilder functions.ImageBuilder
	httpClient   *http.Client
	docker       *docker.Client
}

func New(config *Config) (functions.FaaSDriver, error) {
	defer trace.Trace("")()
	dc, err := docker.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get docker client")
	}

	d := &ofDriver{
		gateway:      strings.TrimRight(config.Gateway, "/"),
		httpClient:   http.DefaultClient,
		imageBuilder: functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, dc),
		docker:       dc,
	}

	return d, nil
}

func (d *ofDriver) Create(f *functions.Function, exec *functions.Exec) error {
	defer trace.Trace("openfaas.Create." + f.ID)()

	image, err := d.imageBuilder.BuildImage("openfaas", f.ID, exec)

	if err != nil {
		return errors.Wrapf(err, "Error building image for function '%s'", f.ID)
	}

	req := requests.CreateFunctionRequest{
		Image:       image,
		Network:     "func_functions",
		Service:     getID(f.ID),
		EnvVars:     map[string]string{},
		Constraints: []string{},
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
	defer trace.Trace("openfaas.Delete." + f.ID)()

	reqBytes, _ := json.Marshal(&requests.DeleteFunctionRequest{FunctionName: getID(f.ID)})
	req, _ := http.NewRequest("DELETE", d.gateway+"/system/functions", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", jsonContentType)

	res, err := d.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Error removing existing function: %s, gateway=%s, functionName=%s\n", err.Error(), d.gateway, f.ID)
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

type ctxAndIn struct {
	Context functions.Context `json:"context"`
	Input   interface{}       `json:"input"`
}

const xStderrHeader = "X-Stderr"

func (d *ofDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		defer trace.Trace("openfaas.run." + e.ID)()

		bytesIn, _ := json.Marshal(ctxAndIn{Context: ctx, Input: in})
		postURL := d.gateway + "/function/" + getID(e.ID)
		res, err := d.httpClient.Post(postURL, jsonContentType, bytes.NewReader(bytesIn))
		if err != nil {
			log.Errorf("Error when sending POST request to %s: %+v", postURL, err)
			return nil, errors.Wrapf(err, "request to OpenFaaS on %s failed", d.gateway)
		}
		defer res.Body.Close()

		log.Debugf("openfaas.run.%s: status code: %v", e.ID, res.StatusCode)
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
