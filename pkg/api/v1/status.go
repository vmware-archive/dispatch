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

// Status status
// swagger:model Status
type Status string

const (

	// StatusINITIALIZED captures enum value "INITIALIZED"
	StatusINITIALIZED Status = "INITIALIZED"

	// StatusCREATING captures enum value "CREATING"
	StatusCREATING Status = "CREATING"

	// StatusREADY captures enum value "READY"
	StatusREADY Status = "READY"

	// StatusUPDATING captures enum value "UPDATING"
	StatusUPDATING Status = "UPDATING"

	// StatusERROR captures enum value "ERROR"
	StatusERROR Status = "ERROR"

	// StatusDELETING captures enum value "DELETING"
	StatusDELETING Status = "DELETING"

	//StatusDELETED captures enum value "DELETED"
	StatusDELETED Status = "DELETED"
)

// NO TESTS

// for schema
var statusEnum []interface{}

func init() {
	var res []Status
	if err := json.Unmarshal([]byte(`["INITIALIZED","CREATING","READY","UPDATING","ERROR","DELETING"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		statusEnum = append(statusEnum, v)
	}
}

func (m Status) validateStatusEnum(path, location string, value Status) error {
	if err := validate.Enum(path, location, value, statusEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this status
func (m Status) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateStatusEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
