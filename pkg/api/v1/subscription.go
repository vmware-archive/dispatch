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

// Subscription subscription
// swagger:model Subscription
type Subscription struct {

	// created time
	// Read Only: true
	CreatedTime int64 `json:"created-time,omitempty"`

	// event type
	// Required: true
	// Max Length: 128
	// Pattern: ^[\w\d\-\.]+$
	EventType *string `json:"event-type"`

	// function
	// Required: true
	// Pattern: ^[\w\d\-]+$
	Function *string `json:"function"`

	// id
	// Read Only: true
	ID strfmt.UUID `json:"id,omitempty"`

	// kind
	// Read Only: true
	// Pattern: ^[\w\d\-]+$
	Kind string `json:"kind,omitempty"`

	// modified time
	// Read Only: true
	ModifiedTime int64 `json:"modified-time,omitempty"`

	// name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*$
	Name *string `json:"name"`

	// secrets
	Secrets []string `json:"secrets"`

	// source type
	// Required: true
	// Max Length: 32
	// Pattern: ^[\w\d\-]+$
	SourceType *string `json:"source-type"`

	// status
	// Read Only: true
	Status Status `json:"status,omitempty"`

	// tags
	Tags []*Tag `json:"tags"`
}

// Validate validates this subscription
func (m *Subscription) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEventType(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateFunction(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
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

	if err := m.validateSecrets(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSourceType(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
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

func (m *Subscription) validateEventType(formats strfmt.Registry) error {
	if err := validate.Required("event-type", "body", m.EventType); err != nil {
		return err
	}

	if err := validate.MaxLength("event-type", "body", string(*m.EventType), 128); err != nil {
		return err
	}

	if err := validate.Pattern("event-type", "body", string(*m.EventType), `^[\w\d\-\.]+$`); err != nil {
		return err
	}
	return nil
}

func (m *Subscription) validateFunction(formats strfmt.Registry) error {

	if err := validate.Required("function", "body", m.Function); err != nil {
		return err
	}

	if err := validate.Pattern("function", "body", string(*m.Function), `^[\w\d\-]+$`); err != nil {
		return err
	}
	return nil
}

func (m *Subscription) validateID(formats strfmt.Registry) error {

	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.FormatOf("id", "body", "uuid", m.ID.String(), formats); err != nil {
		return err
	}
	return nil
}

func (m *Subscription) validateKind(formats strfmt.Registry) error {

	if swag.IsZero(m.Kind) { // not required
		return nil
	}

	if err := validate.Pattern("kind", "body", string(m.Kind), `^[\w\d\-]+$`); err != nil {
		return err
	}
	return nil
}

func (m *Subscription) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := FieldPatternWordNumberDashOnly.Validate("name", *m.Name); err != nil {
		return err
	}

	return nil
}

func (m *Subscription) validateSecrets(formats strfmt.Registry) error {

	if swag.IsZero(m.Secrets) { // not required
		return nil
	}

	return nil
}

func (m *Subscription) validateSourceType(formats strfmt.Registry) error {

	if err := validate.Required("source-type", "body", m.SourceType); err != nil {
		return err
	}

	if err := validate.MaxLength("source-type", "body", string(*m.SourceType), 32); err != nil {
		return err
	}

	if err := validate.Pattern("source-type", "body", string(*m.SourceType), `^[\w\d\-]+$`); err != nil {
		return err
	}
	return nil
}

func (m *Subscription) validateStatus(formats strfmt.Registry) error {

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

func (m *Subscription) validateTags(formats strfmt.Registry) error {

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
func (m *Subscription) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Subscription) UnmarshalBinary(b []byte) error {
	var res Subscription
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
