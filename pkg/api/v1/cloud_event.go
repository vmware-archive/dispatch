///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	"fmt"

	"github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NO TESTS

// CloudEvent cloud event
// swagger:model CloudEvent
type CloudEvent struct {

	// cloud events version
	// Required: true
	CloudEventsVersion *string `json:"cloud-events-version"`

	// content type
	ContentType string `json:"content-type,omitempty"`

	// data
	Data string `json:"data,omitempty"`

	// event id
	// Required: true
	EventID *string `json:"event-id"`

	// event time
	EventTime strfmt.DateTime `json:"event-time,omitempty"`

	// event type
	// Required: true
	// Max Length: 128
	// Pattern: ^[\w\d\-\.]+$
	EventType *string `json:"event-type"`

	// event type version
	EventTypeVersion string `json:"event-type-version,omitempty"`

	// extensions
	Extensions map[string]interface{} `json:"extensions,omitempty"`

	// namespace
	// Required: true
	Namespace *string `json:"namespace"`

	// schema url
	SchemaURL string `json:"schema-url,omitempty"`

	// source id
	// Required: true
	SourceID *string `json:"source-id"`

	// source type
	// Required: true
	SourceType *string `json:"source-type"`
}

// Validate validates this cloud event
func (m *CloudEvent) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCloudEventsVersion(formats); err != nil {
		fmt.Print("validateCloudEventsVersion", err)
		// prop
		res = append(res, err)
	}

	err := m.validateEventID(formats)
	if err != nil {
		fmt.Printf("\nvalue: %v\n", err)
	}

	if err := m.validateEventTime(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateEventType(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateNamespace(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSourceID(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSourceType(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CloudEvent) validateCloudEventsVersion(formats strfmt.Registry) error {
	if err := validate.Required("cloud-events-version", "body", m.CloudEventsVersion); err != nil {
		return err
	}

	return nil
}

func (m *CloudEvent) validateEventID(formats strfmt.Registry) error {
	if err := validate.Required("event-id", "body", m.EventID); err != nil {
		return err
	}

	return nil
}

func (m *CloudEvent) validateEventTime(formats strfmt.Registry) error {
	if swag.IsZero(m.EventTime) { // not required
		return nil
	}

	if err := validate.FormatOf("event-time", "body", "date-time", m.EventTime.String(), formats); err != nil {
		return err
	}
	return nil
}

func (m *CloudEvent) validateEventType(formats strfmt.Registry) error {
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

func (m *CloudEvent) validateNamespace(formats strfmt.Registry) error {
	if err := validate.Required("namespace", "body", m.Namespace); err != nil {
		return err
	}
	return nil
}

func (m *CloudEvent) validateSourceID(formats strfmt.Registry) error {
	if err := validate.Required("source-id", "body", m.SourceID); err != nil {
		return err
	}
	return nil
}

func (m *CloudEvent) validateSourceType(formats strfmt.Registry) error {
	if err := validate.Required("source-type", "body", m.SourceType); err != nil {
		return err
	}
	return nil
}

// MarshalBinary interface implementation
func (m *CloudEvent) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CloudEvent) UnmarshalBinary(b []byte) error {
	var res CloudEvent
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
