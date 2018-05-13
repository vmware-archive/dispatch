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

// Spec spec
// swagger:model Spec
type Spec string

const (

	// SpecCONFIGURE captures enum value "CONFIGURE"
	SpecCONFIGURE Spec = "CONFIGURE"

	// SpecCREATE captures enum value "CREATE"
	SpecCREATE Spec = "CREATE"

	// SpecDELETE captures enum value "DELETE"
	SpecDELETE Spec = "DELETE"
)

// NO TESTS

// for schema
var specEnum []interface{}

func init() {
	var res []Spec
	if err := json.Unmarshal([]byte(`["CONFIGURE","CREATE","DELETE"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		specEnum = append(specEnum, v)
	}
}

func (m Spec) validateSpecEnum(path, location string, value Spec) error {
	if err := validate.Enum(path, location, value, specEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this spec
func (m Spec) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateSpecEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
