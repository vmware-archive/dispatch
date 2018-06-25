///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	"strings"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// NO TESTS

// FieldPattern field pattern
// swagger:model FieldPattern
type FieldPattern struct {

	// regex pattern
	Pattern string `json:"pattern"`

	// translated msg presented to user
	Message string `json:"message"`
}

// FieldPatternName letter, underscore, number and dash
var FieldPatternName = FieldPattern{
	Pattern: `^[\w\d][\w\d\-]*$`,
	Message: "should start with letter or number and may only contain letters, numbers, underscores and dashes",
}

// Validate validates field pattern and return error with translated msg
func (p FieldPattern) Validate(fieldName, field string) error {
	if err := validate.Pattern(fieldName, "body", field, p.Pattern); err != nil {
		return errors.New(err.Code(), strings.Join([]string{fieldName, p.Message}, " "))
	}
	return nil
}
