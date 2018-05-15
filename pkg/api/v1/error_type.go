///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// NO TESTS

// ErrorType error type
// swagger:model ErrorType
type ErrorType string

const (

	// ErrorTypeInputError captures enum value "InputError"
	ErrorTypeInputError ErrorType = "InputError"

	// ErrorTypeFunctionError captures enum value "FunctionError"
	ErrorTypeFunctionError ErrorType = "FunctionError"

	// ErrorTypeSystemError captures enum value "SystemError"
	ErrorTypeSystemError ErrorType = "SystemError"
)

// for schema
var errorTypeEnum []interface{}

func init() {
	var res []ErrorType
	if err := json.Unmarshal([]byte(`["InputError","FunctionError","SystemError"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		errorTypeEnum = append(errorTypeEnum, v)
	}
}

func (m ErrorType) validateErrorTypeEnum(path, location string, value ErrorType) error {
	if err := validate.Enum(path, location, value, errorTypeEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this error type
func (m ErrorType) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateErrorTypeEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
