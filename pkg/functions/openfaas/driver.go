///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package openfaas

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/openfaas/faas/gateway/requests"
	"github.com/pkg/errors"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
)

const (
	jsonContentType = "application/json"
)

type Config struct {
	Gateway string
}

type ofDriver struct {
	gateway    string
	httpClient *http.Client
}

func New(config *Config) functions.FaaSDriver {
	return &ofDriver{
		gateway:    strings.TrimRight(config.Gateway, "/"),
		httpClient: http.DefaultClient,
	}
}

func (d *ofDriver) Create(name string, exec *functions.Exec) error {

	if err := d.Delete(name); err != nil {
		return errors.Wrapf(err, "Failed to cleanup before deploying function '%s'", name)
	}

	req := requests.CreateFunctionRequest{
		Image:       exec.Image,
		Network:     "func_functions",
		Service:     name,
		EnvVars:     map[string]string{},
		Constraints: []string{},
	}

	reqBytes, _ := json.Marshal(&req)
	res, err := d.httpClient.Post(d.gateway+"/system/functions", jsonContentType, bytes.NewReader(reqBytes))
	if err != nil {
		return errors.Wrapf(err, "Error deploying function '%s'", name)
	}
	defer res.Body.Close()

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

func (d *ofDriver) Delete(name string) error {
	reqBytes, _ := json.Marshal(&requests.DeleteFunctionRequest{FunctionName: name})
	req, _ := http.NewRequest("DELETE", d.gateway+"/system/functions", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", jsonContentType)

	res, err := d.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Error removing existing function: %s, gateway=%s, functionName=%s\n", err.Error(), d.gateway, name)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200, 201, 202, 404:
		return nil
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return errors.Errorf("Server returned unexpected status: %v, %s", res.StatusCode, string(bytesOut))
		}
		return errors.Wrapf(err, "Error performing DELETE request, status: %v", res.StatusCode)
	}
}

func (d *ofDriver) GetRunnable(name string) functions.Runnable {
	return func(args map[string]interface{}) (map[string]interface{}, error) {
		bytesIn, _ := json.Marshal(args)
		res, err := d.httpClient.Post(d.gateway+"/function/"+name, jsonContentType, bytes.NewReader(bytesIn))
		if err != nil {
			return nil, errors.Errorf("cannot connect to OpenFaaS on URL: %s", d.gateway)
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case 200:
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, errors.Errorf("cannot read result from OpenFaaS on URL: %s %s", d.gateway, err)
			}
			result := map[string]interface{}{}
			if err := json.Unmarshal(resBytes, &result); err != nil {
				return nil, errors.Errorf("cannot JSON-parse result from OpenFaaS: %s %s", err, string(resBytes))
			}
			return result, nil

		default:
			bytesOut, err := ioutil.ReadAll(res.Body)
			if err == nil {
				return nil, errors.Errorf("Server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
			}
			return nil, errors.Wrapf(err, "Error performing DELETE request, status: %v", res.StatusCode)
		}
	}
}
