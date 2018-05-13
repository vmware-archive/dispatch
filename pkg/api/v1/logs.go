///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NO TESTS

// Logs logs
// swagger:model Logs
type Logs struct {

	// stderr
	// Required: true
	Stderr []string `json:"stderr"`

	// stdout
	// Required: true
	Stdout []string `json:"stdout"`
}

// Validate validates this logs
func (m *Logs) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateStderr(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateStdout(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Logs) validateStderr(formats strfmt.Registry) error {

	if err := validate.Required("stderr", "body", m.Stderr); err != nil {
		return err
	}

	return nil
}

func (m *Logs) validateStdout(formats strfmt.Registry) error {

	if err := validate.Required("stdout", "body", m.Stdout); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Logs) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Logs) UnmarshalBinary(b []byte) error {
	var res Logs
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
