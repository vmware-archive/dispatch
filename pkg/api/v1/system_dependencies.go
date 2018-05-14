///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// NO TESTS

// SystemDependencies system dependencies
// swagger:model SystemDependencies
type SystemDependencies struct {

	// packages
	Packages []*SystemDependency `json:"packages"`
}

// Validate validates this system dependencies
func (m *SystemDependencies) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validatePackages(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SystemDependencies) validatePackages(formats strfmt.Registry) error {

	if swag.IsZero(m.Packages) { // not required
		return nil
	}

	for i := 0; i < len(m.Packages); i++ {

		if swag.IsZero(m.Packages[i]) { // not required
			continue
		}

		if m.Packages[i] != nil {

			if err := m.Packages[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("packages" + "." + strconv.Itoa(i))
				}
				return err
			}

		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *SystemDependencies) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SystemDependencies) UnmarshalBinary(b []byte) error {
	var res SystemDependencies
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
