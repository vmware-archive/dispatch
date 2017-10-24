///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package openwhisk

import (
	"github.com/apache/incubator-openwhisk-client-go/whisk"
	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
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

func (d *wskDriver) Create(id string, exec *functions.Exec) error {
	action := &whisk.Action{
		Name: id,
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

func (d *wskDriver) Delete(id string) error {
	_, err := d.client.Actions.Delete(id)
	return err
}

type ctxAndIn struct {
	Context functions.Context `json:"context"`
	Input   interface{}       `json:"input"`
}

func (d *wskDriver) GetRunnable(id string) functions.Runnable {
	return func(ctx functions.Context, in interface{}) (interface{}, error) {
		result, _, err := d.client.Actions.Invoke(id, ctxAndIn{Context: ctx, Input: in}, true, true)
		if err != nil {
			return nil, err // TODO err should be JSON-serializable and usable (e.g. invalid arg vs runtime error)
		}
		return result, nil
	}
}
