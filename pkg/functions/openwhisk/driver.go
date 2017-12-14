///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package openwhisk

import (
	"github.com/apache/incubator-openwhisk-client-go/whisk"
	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/functions"
)

type wskDriver struct {
	client *whisk.Client
}

type Config struct {
	AuthToken string
	Host      string
	Insecure  bool
}

func New(config *Config) (functions.FaaSDriver, error) {
	baseURL, err := whisk.GetURLBase(config.Host, "/api")
	if err != nil {
		return nil, errors.Wrap(err, "error parsing base URL from API host")
	}
	client, err := whisk.NewClient(nil, &whisk.Config{
		AuthToken: config.AuthToken,
		Host:      config.Host,
		BaseURL:   baseURL,
		Insecure:  config.Insecure,
		Namespace: "",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating openwhisk driver")
	}
	return &wskDriver{client}, nil
}

func (d *wskDriver) Create(f *functions.Function, exec *functions.Exec) error {
	action := &whisk.Action{
		Name: f.ID,
		Exec: &whisk.Exec{
			Code:  &exec.Code,
			Main:  exec.Main,
			Image: exec.Image,
			Kind:  "blackbox",
		},
	}
	_, _, err := d.client.Actions.Insert(action, true)
	return err
}

func (d *wskDriver) Delete(f *functions.Function) error {
	_, err := d.client.Actions.Delete(f.ID)
	return err
}

type ctxAndIn struct {
	Context functions.Context `json:"context"`
	Input   interface{}       `json:"input"`
}

func (d *wskDriver) GetRunnable(e *functions.FunctionExecution) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		result, _, err := d.client.Actions.Invoke(e.ID, ctxAndIn{Context: ctx, Input: in}, true, true)
		if err != nil {
			return nil, err // TODO err should be JSON-serializable and usable (e.g. invalid arg vs runtime error)
		}
		return result, nil
	}
}
