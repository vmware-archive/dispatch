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
	FUNC       = "test/f0"
	SCHEMA_IN  = "test/schemaIn"
	SCHEMA_OUT = "test/schemaOut"
	TEST       = "test"
	VALIDATION = "validation"
	F0         = "f0"
	M1         = "m1"
	M2         = "m2"
	ARGS       = "args"
	TRACE_IN   = "trace-in"
	TRACE_OUT  = "trace-out"
)

func TestRun(t *testing.T) {
	faas := &mocks.FaaSDriver{}
	v := &mocks.Validator{}
	testSchemas := &functions.Schemas{SchemaIn: SCHEMA_IN, SchemaOut: SCHEMA_OUT}

	faas.On("GetRunnable", FUNC).Return(functions.F(f0))
	v.On("GetMiddleware", testSchemas).Return(functions.Middleware(mw0(VALIDATION)))

	testRunner := New(faas, v)

	fn := &functions.Function{
		Name:    FUNC,
		Schemas: testSchemas,
	}
	args := map[string]interface{}{TEST: TEST}

	result, err := testRunner.RunFunction(fn, args)
	faas.AssertExpectations(t)
	v.AssertExpectations(t)
	assert.Nil(t, err)
	expected := map[string]interface{}{
		ARGS:      args,
		TRACE_IN:  []string{VALIDATION, F0},
		TRACE_OUT: []string{F0, VALIDATION},
	}
	assert.Equal(t, expected, result)
}

func f0(args map[string]interface{}) (map[string]interface{}, error) {
	if args == nil {
		return nil, errors.New("nil args")
	}
	traceIn, _ := args[TRACE_IN].([]string)
	return map[string]interface{}{
		ARGS:      args,
		TRACE_IN:  append(traceIn, F0),
		TRACE_OUT: []string{F0},
	}, nil
}

func mw0(n string) functions.Middleware {
	return func(f functions.F) functions.F {
		return func(args map[string]interface{}) (map[string]interface{}, error) {
			if args == nil {
				return nil, errors.New("nil args")
			}

			traceIn, _ := args[TRACE_IN].([]string)
			args[TRACE_IN] = append(traceIn, n)

			result, err := f(args)
			if err != nil {
				return nil, err
			}

			traceOut, _ := result[TRACE_OUT].([]string)
			result[TRACE_OUT] = append(traceOut, n)

			return result, nil
		}
	}
}

func TestCompose(t *testing.T) {
	a0 := map[string]interface{}{TEST: TEST}

	expected := map[string]interface{}{
		ARGS:      a0,
		TRACE_IN:  []string{M1, M2, F0},
		TRACE_OUT: []string{F0, M2, M1},
	}

	result, err := Compose(mw0(M1), mw0(M2))(f0)(a0)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
