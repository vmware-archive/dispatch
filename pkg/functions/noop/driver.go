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
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/trace"
)

// Config for the no-op function driver
type Config struct{}

type noopDriver struct{}

// New is the constructor for the no-op function driver
func New(config *Config) (functions.FaaSDriver, error) {
	return &noopDriver{}, nil
}

// Create creates a function [image] but no actual function
func (d *noopDriver) Create(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return nil
}

// Delete is a no-op
func (d *noopDriver) Delete(ctx context.Context, f *functions.Function) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

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
		return ctxAndIn{Context: ctx, Input: in}, nil
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
