///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package validator

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware/dispatch/pkg/functions"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestImpl_GetMiddleware(t *testing.T) {
	v := New()

	schemas := &functions.Schemas{
		SchemaIn: &spec.Schema{SchemaProps: spec.SchemaProps{
			Properties: map[string]spec.Schema{
				"inputField": spec.Schema{SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				}},
			},
			Required: []string{"inputField"},
		}},
		SchemaOut: &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: map[string]spec.Schema{
					"outputField": spec.Schema{SchemaProps: spec.SchemaProps{
						Type: []string{"string"},
					}},
				},
				Required: []string{"outputField"},
			},
		},
	}
	identity := func(ctx functions.Context, input interface{}) (interface{}, error) {
		return input, nil
	}

	var specPtrNil *spec.Schema

	testCases := []struct {
		name            string
		schemas         *functions.Schemas
		input           map[string]interface{}
		expectedUserErr bool
		expectedFuncErr bool
	}{
		{
			name: "nils",
			schemas: &functions.Schemas{
				SchemaIn:  nil,
				SchemaOut: nil,
			},
			input:           map[string]interface{}{"inputField": "some string"},
			expectedUserErr: false,
			expectedFuncErr: false,
		}, {
			name: "specPtrNils",
			schemas: &functions.Schemas{
				SchemaIn:  specPtrNil,
				SchemaOut: specPtrNil,
			},
			input:           map[string]interface{}{"inputField": "some string"},
			expectedUserErr: false,
			expectedFuncErr: false,
		}, {
			name:            "expectUserErr: nil/empty input",
			schemas:         schemas,
			input:           nil,
			expectedUserErr: true,
			expectedFuncErr: false,
		}, {
			name:            "expectFuncErr",
			schemas:         schemas,
			input:           map[string]interface{}{"inputField": "some string"},
			expectedUserErr: false,
			expectedFuncErr: true,
		}, {
			name:            "expectNoErr",
			schemas:         schemas,
			input:           map[string]interface{}{"inputField": "some string", "outputField": "some string"},
			expectedUserErr: false,
			expectedFuncErr: false,
		}, {
			name:            "expectUserErr",
			schemas:         schemas,
			input:           map[string]interface{}{"inputField": "some string", "outputField": 10},
			expectedUserErr: false,
			expectedFuncErr: true,
		},
	}

	for _, testCase := range testCases {
		log.Debugf("testcase: %s", testCase.name)

		output, err := v.GetMiddleware(testCase.schemas)(identity)(functions.Context{}, testCase.input)

		if !testCase.expectedUserErr && !testCase.expectedFuncErr {
			require.NoError(t, err)
			assert.Equal(t, testCase.input, output)
			continue
		}
		require.Error(t, err)
		_, isUserErr := err.(functions.UserError)
		_, isFuncErr := err.(functions.FunctionError)
		assert.Equal(t, testCase.expectedUserErr, isUserErr)
		assert.Equal(t, testCase.expectedFuncErr, isFuncErr)
	}
}

func TestNumberNoPanic(t *testing.T) {
	var schemaJSON = `
{
    "properties": {
        "name": {
            "type": "string",
            "pattern": "^[A-Za-z]+$",
            "minLength": 1
        },
        "place": {
            "type": "string",
            "pattern": "^[A-Za-z]+$",
            "minLength": 1
        }
    },
    "required": [
        "name"
    ]
}`
	var inputJSON = `{"name": "Ivan"}`
	schema := new(spec.Schema)
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), schema))
	var input map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(inputJSON), &input))
	input["place"] = json.Number("10")

	err := validate.AgainstSchema(schema, input, strfmt.Default)
	assert.Error(t, err)
	log.Debugf("validation error: %s", err)
}
