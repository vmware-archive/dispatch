///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package runner

import (
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/mocks"
)

const (
	testF0        = "test/f0"
	testSchemaIn  = "test/schemaIn"
	testSchemaOut = "test/schemaOut"
	test          = "test"
	validation    = "validation"
	injection     = "injection"
	f0            = "f0"
	m1            = "m1"
	m2            = "m2"
	argsStr       = "args"
	traceInStr    = "trace-in"
	traceOutStr   = "trace-out"
)

func TestRun(t *testing.T) {
	faas := &mocks.FaaSDriver{}
	v := &mocks.Validator{}
	injector := &mocks.SecretInjector{}
	testSchemas := &functions.Schemas{SchemaIn: testSchemaIn, SchemaOut: testSchemaOut}
	fe := &functions.FunctionExecution{
		Context: functions.Context{},
		ID:      "",
		Name:    testF0,
		Schemas: testSchemas,
		Secrets: []string{},
		Cookie:  "cookie",
	}

	faas.On("GetRunnable", fe).Return(functions.Runnable(runnable0))
	v.On("GetMiddleware", testSchemas).Return(functions.Middleware(mw0(validation)))
	injector.On("GetMiddleware", []string{}, "cookie").Return(functions.Middleware(mw0(injection)))

	testRunner := New(&Config{faas, v, injector})

	fn := &functions.FunctionExecution{
		Context: functions.Context{},
		Name:    testF0,
		Schemas: testSchemas,
		Secrets: []string{},
		Cookie:  "cookie",
	}
	args := map[string]interface{}{test: test}

	result, err := testRunner.Run(fn, args)
	faas.AssertExpectations(t)
	v.AssertExpectations(t)
	assert.Nil(t, err)
	expected := map[string]interface{}{
		argsStr:     args,
		traceInStr:  []string{validation, injection, f0},
		traceOutStr: []string{f0, injection, validation},
	}
	assert.Equal(t, expected, result)
}

func runnable0(ctx functions.Context, in interface{}) (interface{}, error) {
	args := in.(map[string]interface{})
	if args == nil {
		return nil, errors.New("nil args")
	}
	traceIn, _ := args[traceInStr].([]string)
	return map[string]interface{}{
		argsStr:     args,
		traceInStr:  append(traceIn, f0),
		traceOutStr: []string{f0},
	}, nil
}

func mw0(n string) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, in interface{}) (interface{}, error) {
			args := in.(map[string]interface{})

			traceIn, _ := args[traceInStr].([]string)
			args[traceInStr] = append(traceIn, n)

			out, err := f(ctx, in)
			if err != nil {
				return nil, err
			}
			result := out.(map[string]interface{})

			traceOut, _ := result[traceOutStr].([]string)
			result[traceOutStr] = append(traceOut, n)

			return result, nil
		}
	}
}

func TestCompose(t *testing.T) {
	a0 := map[string]interface{}{test: test}

	expected := map[string]interface{}{
		argsStr:     a0,
		traceInStr:  []string{m1, m2, f0},
		traceOutStr: []string{f0, m2, m1},
	}

	result, err := Compose(mw0(m1), mw0(m2))(runnable0)(functions.Context{}, a0)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
