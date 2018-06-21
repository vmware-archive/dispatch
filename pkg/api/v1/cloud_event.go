///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package v1

import (
	"encoding/json"
	"fmt"

	"github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NO TESTS

// CloudEvent cloud event, implemented based on: https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/spec.md
// swagger:model CloudEvent
type CloudEvent struct {
	// event type
	// Required: true
	// Max Length: 128
	// Pattern: ^[\w\d\-\.]+$
	EventType string `json:"eventType"`

	// event type version
	EventTypeVersion string `json:"eventTypeVersion,omitempty"`

	// cloud events version
	// Required: true
	CloudEventsVersion string `json:"cloudEventsVersion"`

	// source
	// Required: true
	Source string `json:"source"`

	// event id
	// Required: true
	EventID string `json:"eventID"`

	// event time
	EventTime strfmt.DateTime `json:"eventTime,omitempty"`

	// schema url
	SchemaURL string `json:"schemaURL,omitempty"`

	// content type
	ContentType string `json:"contentType,omitempty"`

	// extensions
	Extensions map[string]interface{} `json:"extensions,omitempty"`

	// data
	Data json.RawMessage `json:"data,omitempty"`
}

// Validate validates this cloud event
func (m *CloudEvent) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCloudEventsVersion(formats); err != nil {
		fmt.Print("validateCloudEventsVersion", err)
		// prop
		res = append(res, err)
	}

	if err := m.validateEventID(formats); err != nil {
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

	if err := m.validateSource(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CloudEvent) validateEventType(formats strfmt.Registry) error {
	if err := validate.RequiredString("eventType", "body", m.EventType); err != nil {
		return err
	}

	if err := validate.MaxLength("eventType", "body", m.EventType, 128); err != nil {
		return err
	}

	if err := validate.Pattern("eventType", "body", m.EventType, `^[\w\d\-\.]+$`); err != nil {
		return err
	}
	return nil
}

func (m *CloudEvent) validateCloudEventsVersion(formats strfmt.Registry) error {
	if err := validate.RequiredString("cloudEventsVersion", "body", m.CloudEventsVersion); err != nil {
		return err
	}

	return nil
}

func (m *CloudEvent) validateEventID(formats strfmt.Registry) error {
	if err := validate.RequiredString("eventID", "body", m.EventID); err != nil {
		return err
	}

	return nil
}

func (m *CloudEvent) validateEventTime(formats strfmt.Registry) error {
	if swag.IsZero(m.EventTime) { // not required
		return nil
	}

	if err := validate.FormatOf("eventTime", "body", "date-time", m.EventTime.String(), formats); err != nil {
		return err
	}
	return nil
}

func (m *CloudEvent) validateSource(formats strfmt.Registry) error {
	if err := validate.RequiredString("source", "body", m.Source); err != nil {
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
