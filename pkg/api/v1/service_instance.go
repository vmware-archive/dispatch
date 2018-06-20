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

// ServiceInstance service instance
// swagger:model ServiceInstance
type ServiceInstance struct {

	// binding
	Binding *ServiceBinding `json:"binding,omitempty"`

	// created time
	CreatedTime int64 `json:"createdTime,omitempty"`

	// id
	ID strfmt.UUID `json:"id,omitempty"`

	// kind
	// Read Only: true
	// Pattern: ^[\w\d\-]+$
	Kind string `json:"kind,omitempty"`

	// name
	// Required: true
	// Pattern: ^[\w\d][\w\d\-]*$
	Name *string `json:"name"`

	// parameters
	Parameters interface{} `json:"parameters,omitempty"`

	// reason
	Reason []string `json:"reason"`

	// secret parameters
	SecretParameters []string `json:"secretParameters"`

	// service class
	// Required: true
	// Pattern: ^[\w\d\-]+$
	ServiceClass *string `json:"serviceClass"`

	// service plan
	// Required: true
	// Pattern: ^[\w\d\-]+$
	ServicePlan *string `json:"servicePlan"`

	// status
	Status Status `json:"status,omitempty"`

	// tags
	Tags []*Tag `json:"tags"`
}

// Validate validates this service instance
func (m *ServiceInstance) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBinding(formats); err != nil {
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

	if err := m.validateSecretParameters(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateServiceClass(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateServicePlan(formats); err != nil {
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

func (m *ServiceInstance) validateBinding(formats strfmt.Registry) error {

	if swag.IsZero(m.Binding) { // not required
		return nil
	}

	if m.Binding != nil {

		if err := m.Binding.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("binding")
			}
			return err
		}

	}

	return nil
}

func (m *ServiceInstance) validateID(formats strfmt.Registry) error {

	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.FormatOf("id", "body", "uuid", m.ID.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *ServiceInstance) validateKind(formats strfmt.Registry) error {

	if swag.IsZero(m.Kind) { // not required
		return nil
	}

	if err := validate.Pattern("kind", "body", string(m.Kind), `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (m *ServiceInstance) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := FieldPatternWordNumberDashOnly.Validate("name", *m.Name); err != nil {
		return err
	}

	return nil
}

func (m *ServiceInstance) validateReason(formats strfmt.Registry) error {

	if swag.IsZero(m.Reason) { // not required
		return nil
	}

	return nil
}

func (m *ServiceInstance) validateSecretParameters(formats strfmt.Registry) error {

	if swag.IsZero(m.SecretParameters) { // not required
		return nil
	}

	return nil
}

func (m *ServiceInstance) validateServiceClass(formats strfmt.Registry) error {

	if err := validate.Required("serviceClass", "body", m.ServiceClass); err != nil {
		return err
	}

	if err := validate.Pattern("serviceClass", "body", string(*m.ServiceClass), `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (m *ServiceInstance) validateServicePlan(formats strfmt.Registry) error {

	if err := validate.Required("servicePlan", "body", m.ServicePlan); err != nil {
		return err
	}

	if err := validate.Pattern("servicePlan", "body", string(*m.ServicePlan), `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (m *ServiceInstance) validateStatus(formats strfmt.Registry) error {

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

func (m *ServiceInstance) validateTags(formats strfmt.Registry) error {

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
func (m *ServiceInstance) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ServiceInstance) UnmarshalBinary(b []byte) error {
	var res ServiceInstance
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
