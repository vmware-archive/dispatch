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
	"github.com/go-openapi/validate"
)

// NO TESTS

// EventDriverType driver type
// swagger:model EventDriverType
type EventDriverType struct {

	// built in
	// Read Only: true
	BuiltIn *bool `json:"built-in,omitempty"`

	// config
	Config []*Config `json:"config"`

	// created time
	// Read Only: true
	CreatedTime int64 `json:"created-time,omitempty"`

	// id
	// Read Only: true
	ID strfmt.UUID `json:"id,omitempty"`

	// image
	// Required: true
	Image *string `json:"image"`

	// kind
	// Read Only: true
	// Pattern: ^[\w\d\-]+$
	Kind string `json:"kind,omitempty"`

	// modified time
	// Read Only: true
	ModifiedTime int64 `json:"modified-time,omitempty"`

	// name
	// Required: true
	// Max Length: 32
	// Pattern: ^[\w\d][\w\d\-]*$
	Name *string `json:"name"`

	// tags
	Tags []*Tag `json:"tags"`
}

// Validate validates this driver type
func (m *EventDriverType) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateConfig(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateImage(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateKind(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateTags(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *EventDriverType) validateConfig(formats strfmt.Registry) error {
	if swag.IsZero(m.Config) { // not required
		return nil
	}

	for i := 0; i < len(m.Config); i++ {

		if swag.IsZero(m.Config[i]) { // not required
			continue
		}

		if m.Config[i] != nil {

			if err := m.Config[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("config" + "." + strconv.Itoa(i))
				}
				return err
			}

		}

	}

	return nil
}

func (m *EventDriverType) validateID(formats strfmt.Registry) error {
	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.FormatOf("id", "body", "uuid", m.ID.String(), formats); err != nil {
		return err
	}
	return nil
}

func (m *EventDriverType) validateImage(formats strfmt.Registry) error {
	if err := validate.Required("image", "body", m.Image); err != nil {
		return err
	}
	return nil
}

func (m *EventDriverType) validateKind(formats strfmt.Registry) error {

	if swag.IsZero(m.Kind) { // not required
		return nil
	}

	if err := validate.Pattern("kind", "body", string(m.Kind), `^[\w\d\-]+$`); err != nil {
		return err
	}
	return nil
}

func (m *EventDriverType) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := validate.MaxLength("name", "body", string(*m.Name), 32); err != nil {
		return err
	}

	if err := FieldPatternWordNumberDashOnly.Validate("name", *m.Name); err != nil {
		return err
	}
	return nil
}

func (m *EventDriverType) validateTags(formats strfmt.Registry) error {

	if swag.IsZero(m.Tags) { // not required
		return nil
	}

	for i := 0; i < len(m.Tags); i++ {

		if swag.IsZero(m.Tags[i]) { // not required
			continue
		}

		if m.Tags[i] != nil {

			if err := m.Tags[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("tags" + "." + strconv.Itoa(i))
				}
				return err
			}

		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *EventDriverType) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *EventDriverType) UnmarshalBinary(b []byte) error {
	var res EventDriverType
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
