///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// RuntimeDependencies runtime dependencies
// swagger:model RuntimeDependencies
type RuntimeDependencies struct {

	// format
	Format string `json:"format,omitempty"`

	// manifest
	Manifest string `json:"manifest,omitempty"`
}

// Validate validates this runtime dependencies
func (m *RuntimeDependencies) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFormat(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var runtimeDependenciesTypeFormatPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["pip","pip3","npm"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		runtimeDependenciesTypeFormatPropEnum = append(runtimeDependenciesTypeFormatPropEnum, v)
	}
}

const (

	// RuntimeDependenciesFormatPip captures enum value "pip"
	RuntimeDependenciesFormatPip string = "pip"

	// RuntimeDependenciesFormatPip3 captures enum value "pip3"
	RuntimeDependenciesFormatPip3 string = "pip3"

	// RuntimeDependenciesFormatNpm captures enum value "npm"
	RuntimeDependenciesFormatNpm string = "npm"
)

// prop value enum
func (m *RuntimeDependencies) validateFormatEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, runtimeDependenciesTypeFormatPropEnum); err != nil {
		return err
	}
	return nil
}

func (m *RuntimeDependencies) validateFormat(formats strfmt.Registry) error {

	if swag.IsZero(m.Format) { // not required
		return nil
	}

	// value enum
	if err := m.validateFormatEnum("format", "body", m.Format); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *RuntimeDependencies) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RuntimeDependencies) UnmarshalBinary(b []byte) error {
	var res RuntimeDependencies
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
