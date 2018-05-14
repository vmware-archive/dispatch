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

// ServiceBinding service binding
// swagger:model ServiceBinding
type ServiceBinding struct {

	// binding secret
	BindingSecret string `json:"bindingSecret,omitempty"`

	// created time
	CreatedTime int64 `json:"createdTime,omitempty"`

	// parameters
	Parameters interface{} `json:"parameters,omitempty"`

	// reason
	Reason []string `json:"reason"`

	// secret parameters
	SecretParameters []string `json:"secretParameters"`

	// status
	Status Status `json:"status,omitempty"`
}

// Validate validates this service binding
func (m *ServiceBinding) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateReason(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSecretParameters(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ServiceBinding) validateReason(formats strfmt.Registry) error {

	if swag.IsZero(m.Reason) { // not required
		return nil
	}

	return nil
}

func (m *ServiceBinding) validateSecretParameters(formats strfmt.Registry) error {

	if swag.IsZero(m.SecretParameters) { // not required
		return nil
	}

	return nil
}

func (m *ServiceBinding) validateStatus(formats strfmt.Registry) error {

	if swag.IsZero(m.Status) { // not required
		return nil
	}

	if err := m.Status.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("status")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ServiceBinding) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ServiceBinding) UnmarshalBinary(b []byte) error {
	var res ServiceBinding
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
