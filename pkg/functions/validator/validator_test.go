///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package validator

import (
	"testing"

	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestImpl_GetMiddleware(t *testing.T) {
	v := New()

	schemas := &functions.Schemas{
		SchemaIn: &spec.Schema{SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"inputField": spec.Schema{SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				}},
			},
			Required: []string{"inputField"},
		}},
		SchemaOut: &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
				Properties: map[string]spec.Schema{
					"outputField": spec.Schema{SchemaProps: spec.SchemaProps{
						Type: []string{"string"},
					}},
				},
				Required: []string{"outputField"},
			},
		},
	}
	identity := func(input map[string]interface{}) (map[string]interface{}, error) {
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
			name:            "expectUserErr",
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
		},
	}

	for _, testCase := range testCases {
		log.Debugf("testcase: %s", testCase.name)

		output, err := v.GetMiddleware(testCase.schemas)(identity)(testCase.input)

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
