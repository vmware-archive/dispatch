///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// NO TESTS

// ServicePlanSchema service plan schema
// swagger:model ServicePlanSchema
type ServicePlanSchema struct {

	// bind
	Bind interface{} `json:"bind,omitempty"`

	// create
	Create interface{} `json:"create,omitempty"`

	// update
	Update interface{} `json:"update,omitempty"`
}

// Validate validates this service plan schema
func (m *ServicePlanSchema) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// MarshalBinary interface implementation
func (m *ServicePlanSchema) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ServicePlanSchema) UnmarshalBinary(b []byte) error {
	var res ServicePlanSchema
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
