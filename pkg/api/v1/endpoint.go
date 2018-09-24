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

// Endpoint Endpoint
// swagger:model Endpoint
type Endpoint struct {
	// meta
	Meta

	// enable Cross-Origin Resource Sharing (CORS)
	Cors bool `json:"cors,omitempty"`

	// a easy way to disable an Endpoint without deleting it.
	Enabled bool `json:"enabled,omitempty"`

	// the name of the function associated with
	// Required: true
	Function string `json:"function"`

	// a list of domain names that point to the Endpoint
	Hosts []string `json:"hosts"`

	// a list of HTTP/S methods that point to the Endpoint
	Methods []string `json:"methods"`

	// a list of support protocols (i.e. http, https)
	Protocols []string `json:"protocols"`

	// status
	Status Status `json:"status,omitempty"`

	// a list of URIs prefixes that point to the Endpoint
	Uris []string `json:"uris"`
}

// NewEndpoint returns a default new endpoint
func NewEndpoint() *Endpoint {
	return &Endpoint{
		Meta: Meta{
			Kind: EndpointKind,
		},
	}
}

// Validate validates this Endpoint
func (m *Endpoint) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFunction(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateHosts(formats); err != nil {
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

	if err := m.validateMethods(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateProtocols(formats); err != nil {
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

	if err := m.validateUris(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Endpoint) validateFunction(formats strfmt.Registry) error {

	if err := validate.Required("function", "body", m.Function); err != nil {
		return err
	}

	return nil
}

func (m *Endpoint) validateHosts(formats strfmt.Registry) error {

	if swag.IsZero(m.Hosts) { // not required
		return nil
	}

	return nil
}

func (m *Endpoint) validateID(formats strfmt.Registry) error {

	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.FormatOf("id", "body", "uuid", m.ID.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Endpoint) validateKind(formats strfmt.Registry) error {

	if swag.IsZero(m.Kind) { // not required
		return nil
	}

	if err := validate.Pattern("kind", "body", string(m.Kind), `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (m *Endpoint) validateMethods(formats strfmt.Registry) error {

	if swag.IsZero(m.Methods) { // not required
		return nil
	}

	return nil
}

func (m *Endpoint) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := FieldPatternName.Validate("name", m.Name); err != nil {
		return err
	}

	return nil
}

func (m *Endpoint) validateProtocols(formats strfmt.Registry) error {

	if swag.IsZero(m.Protocols) { // not required
		return nil
	}

	return nil
}

func (m *Endpoint) validateStatus(formats strfmt.Registry) error {

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

func (m *Endpoint) validateTags(formats strfmt.Registry) error {

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

func (m *Endpoint) validateUris(formats strfmt.Registry) error {

	if swag.IsZero(m.Uris) { // not required
		return nil
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Endpoint) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Endpoint) UnmarshalBinary(b []byte) error {
	var res Endpoint
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
