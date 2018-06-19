///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// NO TESTS

// FieldPattern field pattern
// wsagger:model FieldPattern
type FieldPattern struct {

	// regex pattern
	Pattern string `json:"pattern"`

	// translated msg presented to user
	Message string `json:"message"`
}

// FieldPatternWordNumberDashOnly letter, underscore, number and dash
var FieldPatternWordNumberDashOnly = FieldPattern{
	Pattern: `^[\w\d][\w\d\-]*$`,
	Message: " should start with letter or number and may only contain letters, numbers, underscores and dashes",
}

// FieldPatternLetterNumberDashOnly letter, number and dash
var FieldPatternLetterNumberDashOnly = FieldPattern{
	Pattern: `^[a-zA-Z0-9][a-zA-Z0-9\-]*$`,
	Message: " should start with letter or number and may only contain letters, numbers and dashes",
}

// Validate validates field pattern and return error with translated msg
func (p FieldPattern) Validate(fieldName, field string) error {
	if err := validate.Pattern(fieldName, "body", field, p.Pattern); err != nil {
		return errors.New(err.Code(), fieldName+p.Message)
	}
	return nil
}
