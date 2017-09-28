///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package validator

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
)

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

	mw := v.GetMiddleware(schemas)
	f := func(input map[string]interface{}) (map[string]interface{}, error) {
		return input, nil
	}

	input := map[string]interface{}{"inputField": "some string"}

	output, err := mw(f)(input)
	assert.Nil(t, output) // because of output validation error
	fe, ok := err.(functions.FunctionError)
	assert.True(t, ok)
	assert.NotNil(t, fe.AsFunctionErrorObject())
}
