///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package validator

import (
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/functions"
)

type schemaValidator struct {
}

// New creates a schema validator
func New() functions.Validator {
	return &schemaValidator{}
}

type inputError struct {
	Err error `json:"err"`
}

func (err *inputError) Error() string {
	return err.Err.Error()
}

func (err *inputError) AsUserErrorObject() interface{} {
	return err
}

type outputError struct {
	Err error `json:"err"`
}

func (err *outputError) Error() string {
	return err.Err.Error()
}

func (err *outputError) AsFunctionErrorObject() interface{} {
	return err
}

func (*schemaValidator) GetMiddleware(schemas *functions.Schemas) functions.Middleware {
	return func(f functions.Runnable) functions.Runnable {
		return func(ctx functions.Context, input interface{}) (interface{}, error) {
			if schema, ok := schemas.SchemaIn.(*spec.Schema); ok {
				if schema != nil {
					if err := validate.AgainstSchema(schema, input, strfmt.Default); err != nil {
						return nil, &inputError{err}
					}
				}
			} else {
				log.Warnf("Unknown schema impl: %v", schema)
			}
			output, err := f(ctx, input)
			if err != nil {
				return nil, err
			}
			if schema, ok := schemas.SchemaOut.(*spec.Schema); ok {
				if schema != nil {
					if err := validate.AgainstSchema(schema, output, strfmt.Default); err != nil {
						return nil, &outputError{err}
					}
				}
			} else {
				log.Warnf("Unknown schema impl: %v", schema)
			}
			return output, nil
		}
	}
}
