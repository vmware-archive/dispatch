///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package validator

import (
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
)

type schemaValidator struct {
}

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
		return func(input map[string]interface{}) (map[string]interface{}, error) {
			if schemas.SchemaIn != nil {
				if err := validate.AgainstSchema(schemas.SchemaIn.(*spec.Schema), input, strfmt.Default); err != nil {
					return nil, &inputError{err}
				}
			}
			output, err := f(input)
			if err != nil {
				return nil, err
			}
			if schemas.SchemaOut != nil {
				if err := validate.AgainstSchema(schemas.SchemaOut.(*spec.Schema), output, strfmt.Default); err != nil {
					return nil, &outputError{err}
				}
			}
			return output, nil
		}
	}
}
