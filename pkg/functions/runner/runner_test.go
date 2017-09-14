///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package runner

import (
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/mocks"
)

const (
	testFnId      = "some_uuid"
	testF0        = "test/f0"
	testSchemaIn  = "test/schemaIn"
	testSchemaOut = "test/schemaOut"
	test          = "test"
	validation    = "validation"
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
	testSchemas := &functions.Schemas{SchemaIn: testSchemaIn, SchemaOut: testSchemaOut}

	faas.On("GetRunnable", testFnId).Return(functions.Runnable(runnable0))
	v.On("GetMiddleware", testSchemas).Return(functions.Middleware(mw0(validation)))

	testRunner := New(&Config{faas, v})

	fn := &functions.Function{
		ID:      testFnId,
		Name:    testF0,
		Schemas: testSchemas,
	}
	args := map[string]interface{}{test: test}

	result, err := testRunner.Run(fn, args)
	faas.AssertExpectations(t)
	v.AssertExpectations(t)
	assert.Nil(t, err)
	expected := map[string]interface{}{
		argsStr:     args,
		traceInStr:  []string{validation, f0},
		traceOutStr: []string{f0, validation},
	}
	assert.Equal(t, expected, result)
}

func runnable0(args map[string]interface{}) (map[string]interface{}, error) {
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
		return func(args map[string]interface{}) (map[string]interface{}, error) {
			if args == nil {
				return nil, errors.New("nil args")
			}

			traceIn, _ := args[traceInStr].([]string)
			args[traceInStr] = append(traceIn, n)

			result, err := f(args)
			if err != nil {
				return nil, err
			}

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

	result, err := Compose(mw0(m1), mw0(m2))(runnable0)(a0)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
