///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package events

import (
	"encoding/json"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/satori/go.uuid"
)

// NO TESTS

const (
	// CloudEventsVersion defines version of CloudEvent specification used in Dispatch
	CloudEventsVersion = "0.1"
)

// NewCloudEventWithDefaults creates new copy of CloudEvent struct, using reasonable defaults for all
// mandatory attributes, requiring only eventType to be explicitly specified.
func NewCloudEventWithDefaults(eventType string) CloudEvent {
	return CloudEvent{
		EventType:          eventType,
		CloudEventsVersion: CloudEventsVersion,
		Source:             "dispatch",
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Now(),
	}
}

// CloudEvent structure implements CloudEvent spec:
// https://github.com/cloudevents/spec/blob/b0124528486d3f6b9a247cadd68d91b44b3d3ef4/spec.md
type CloudEvent struct {
	// Mandatory, e.g. "user.created"
	EventType string `json:"eventType" validate:"required,max=128,eventtype"`
	// Optional, e.g. "VMODL6.5"
	EventTypeVersion string `json:"eventTypeVersion,omitempty" validate:"omitempty,min=1"`
	// Mandatory, fixed to "0.1"
	CloudEventsVersion string `json:"cloudEventsVersion" validate:"eq=0.1"`
	// Mandatory, e.g. "vcenter1.corp.local"
	Source string `json:"source" validate:"required,max=64"`
	// Mandatory, e.g. UUID or "43252363". Must be unique for this Source
	EventID string `json:"eventID" validate:"required"`
	// Optional, Timestamp in RFC 3339 format, e.g. "1985-04-12T23:20:50.52Z"
	EventTime time.Time `json:"eventTime,omitempty" validate:"-"`
	// Optional, if specified must be a valid URI
	SchemaURL string `json:"schemaURL,omitempty" validate:"omitempty,uri"`
	// Optional, if specified must be a valid mime type, e.g. "application/json"
	ContentType string `json:"contentType,omitempty" validate:"omitempty,min=1"`
	// Optional, key-value dictionary for use by Dispatch
	Extensions CloudEventExtensions `json:"extensions,omitempty" validate:"omitempty,min=1"`

	// Event payload
	Data json.RawMessage `json:"data" validate:"omitempty"`
}

// CloudEventExtensions holds attributes for CloudEvent that are not part of the standard.
type CloudEventExtensions map[string]interface{}

// InjectSpan injects OpenTracing Span into a CloudEvent.
func (e *CloudEvent) InjectSpan(span opentracing.Span) error {
	return span.Tracer().Inject(span.Context(), opentracing.TextMap, e.Extensions)
}

// ExtractSpan extracts OpenTracing Span from a CloudEvent.
func (e *CloudEvent) ExtractSpan() (opentracing.SpanContext, error) {
	return opentracing.GlobalTracer().Extract(opentracing.TextMap, e.Extensions)
}

// DefaultTopic returns a default representation of topic for messaging purposes.
func (e *CloudEvent) DefaultTopic() string {
	return e.EventType
}

// ForeachKey conforms to the opentracing TextMapReader interface.
func (ex CloudEventExtensions) ForeachKey(handler func(key, val string) error) error {
	for k, val := range ex {
		v, ok := val.(string)
		if !ok {
			continue
		}
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Set implements Set() of opentracing.TextMapWriter.
func (ex CloudEventExtensions) Set(key, val string) {
	ex[key] = val
}

func truncStr(str string, num int) string {
	result := str
	if len(str) > num {
		result = str[0:num]
	}
	return result
}
