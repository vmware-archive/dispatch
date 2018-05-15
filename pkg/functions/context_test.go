///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api/v1"
)

func TestContext_Logs(t *testing.T) {
	ctx := Context{}
	ctx.ReadLogs(bytes.NewReader([]byte("foo\nbar\n")), bytes.NewReader([]byte("foobar\nbarfoo\n")))
	logs := v1.Logs{
		Stderr: []string{"foo", "bar"},
		Stdout: []string{"foobar", "barfoo"},
	}
	assert.Equal(t, logs, ctx.Logs())
}

func TestContext_LogsBS(t *testing.T) {
	ctx := Context{}
	ctx["logs"] = &struct {
		bs string
	}{}
	logs := ctx.Logs()
	assert.Len(t, logs.Stderr, 0)
	assert.Len(t, logs.Stdout, 0)

	ctx["logs"] = map[string]interface{}{
		"stderr": []interface{}{"ln1", "ln2"},
		"stdout": []interface{}{"ln3", "ln4"},
	}
	logs = v1.Logs{
		Stderr: []string{"ln1", "ln2"},
		Stdout: []string{"ln3", "ln4"},
	}
	assert.Equal(t, logs, ctx.Logs())
}

func TestContext_LogsNil(t *testing.T) {
	ctx := Context{}
	assert.Len(t, ctx.Logs().Stderr, 0)
	assert.Len(t, ctx.Logs().Stdout, 0)
}

func TestContext_GetError(t *testing.T) {
	m := make(map[string]interface{})
	m["message"] = "test"
	m["stacktrace"] = []string{"stacktrace", "stacktrace2"}
	m["type"] = v1.ErrorTypeFunctionError

	ctx := Context{}
	ctx[ErrorKey] = m

	assert.Equal(t, "test", *ctx.GetError().Message)
	assert.Equal(t, []string{"stacktrace", "stacktrace2"}, ctx.GetError().Stacktrace)
	assert.Equal(t, v1.ErrorTypeFunctionError, ctx.GetError().Type)
}

func TestContext_GetErrorBS(t *testing.T) {
	ctx := Context{}
	ctx[ErrorKey] = &struct {
		bs string
	}{}

	assert.Nil(t, ctx.GetError())
}

func TestContext_GetErrorInvalidError(t *testing.T) {
	m := make(map[string]interface{})
	m["message"] = "test"
	m["stacktrace"] = false
	m["type"] = v1.ErrorTypeFunctionError

	ctx := Context{}
	ctx[ErrorKey] = m

	assert.Nil(t, ctx.GetError())
}

func TestContext_GetErrorMissingFields(t *testing.T) {
	m := make(map[string]interface{})
	m["message"] = "test"
	m["type"] = v1.ErrorTypeFunctionError

	ctx := Context{}
	ctx[ErrorKey] = m

	assert.Equal(t, "test", *ctx.GetError().Message)
	assert.Len(t, ctx.GetError().Stacktrace, 0)
	assert.Equal(t, v1.ErrorTypeFunctionError, ctx.GetError().Type)
}

func TestContext_GetErrorEmpty(t *testing.T) {
	var m map[string]interface{}

	ctx := Context{}
	ctx[ErrorKey] = m

	assert.Nil(t, ctx.GetError().Message)
	assert.Len(t, ctx.GetError().Stacktrace, 0)
	assert.EqualValues(t, "", ctx.GetError().Type)
}

func TestContext_GetErrorNil(t *testing.T) {
	ctx := Context{}
	assert.Nil(t, ctx.GetError())
}
