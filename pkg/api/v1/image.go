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

// Image image
// swagger:model Image
type Image struct {

	// base image name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*$
	BaseImageName *string `json:"baseImageName"`

	// created time
	CreatedTime int64 `json:"createdTime,omitempty"`

	// docker Url
	// Read Only: true
	DockerURL string `json:"dockerUrl,omitempty"`

	// groups
	Groups []string `json:"groups"`

	// id
	ID strfmt.UUID `json:"id,omitempty"`

	// kind
	// Read Only: true
	// Pattern: ^[\w\d\-]+$
	Kind string `json:"kind,omitempty"`

	// language
	Language string `json:"language,omitempty"`

	// name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*$
	Name *string `json:"name"`

	// reason
	Reason []string `json:"reason"`

	// runtime dependencies
	RuntimeDependencies *RuntimeDependencies `json:"runtimeDependencies,omitempty"`

	// spec
	Spec Spec `json:"spec,omitempty"`

	// status
	Status Status `json:"status,omitempty"`

	// system dependencies
	SystemDependencies *SystemDependencies `json:"systemDependencies,omitempty"`

	// tags
	Tags []*Tag `json:"tags"`
}

// Validate validates this image
func (m *Image) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBaseImageName(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateGroups(formats); err != nil {
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

	if err := m.validateReason(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateRuntimeDependencies(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSpec(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSystemDependencies(formats); err != nil {
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

func (m *Image) validateBaseImageName(formats strfmt.Registry) error {

	if err := validate.Required("baseImageName", "body", m.BaseImageName); err != nil {
		return err
	}

	if err := FieldPatternName.Validate("baseImageName", *m.Name); err != nil {
		return err
	}

	return nil
}

func (m *Image) validateGroups(formats strfmt.Registry) error {

	if swag.IsZero(m.Groups) { // not required
		return nil
	}

	return nil
}

func (m *Image) validateID(formats strfmt.Registry) error {

	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.FormatOf("id", "body", "uuid", m.ID.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Image) validateKind(formats strfmt.Registry) error {

	if swag.IsZero(m.Kind) { // not required
		return nil
	}

	if err := validate.Pattern("kind", "body", string(m.Kind), `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (m *Image) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := FieldPatternName.Validate("name", *m.Name); err != nil {
		return err
	}

	return nil
}

func (m *Image) validateReason(formats strfmt.Registry) error {

	if swag.IsZero(m.Reason) { // not required
		return nil
	}

	return nil
}

func (m *Image) validateRuntimeDependencies(formats strfmt.Registry) error {

	if swag.IsZero(m.RuntimeDependencies) { // not required
		return nil
	}

	if m.RuntimeDependencies != nil {

		if err := m.RuntimeDependencies.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("runtimeDependencies")
			}
			return err
		}

	}

	return nil
}

func (m *Image) validateSpec(formats strfmt.Registry) error {

	if swag.IsZero(m.Spec) { // not required
		return nil
	}

	if err := m.Spec.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("spec")
		}
		return err
	}

	return nil
}

func (m *Image) validateStatus(formats strfmt.Registry) error {

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

func (m *Image) validateSystemDependencies(formats strfmt.Registry) error {

	if swag.IsZero(m.SystemDependencies) { // not required
		return nil
	}

	if m.SystemDependencies != nil {

		if err := m.SystemDependencies.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("systemDependencies")
			}
			return err
		}

	}

	return nil
}

func (m *Image) validateTags(formats strfmt.Registry) error {

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
func (m *Image) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Image) UnmarshalBinary(b []byte) error {
	var res Image
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
