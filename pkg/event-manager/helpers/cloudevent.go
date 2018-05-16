///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package helpers

import (
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/events"
)

// NO TESTS

// CloudEventFromAPI creates CloudEvent struct from API model
func CloudEventFromAPI(e *v1.CloudEvent) *events.CloudEvent {
	if e == nil {
		return nil
	}
	return &events.CloudEvent{
		Namespace:          *e.Namespace,
		EventType:          *e.EventType,
		EventTypeVersion:   e.EventTypeVersion,
		CloudEventsVersion: *e.CloudEventsVersion,
		SourceType:         *e.SourceType,
		SourceID:           *e.SourceID,
		EventID:            *e.EventID,
		EventTime:          time.Time(e.EventTime),
		SchemaURL:          e.SchemaURL,
		ContentType:        e.ContentType,
		Extensions:         events.CloudEventExtensions(e.Extensions),
		Data:               e.Data,
	}
}

// CloudEventToAPI creates API model from CloudEvent struct
func CloudEventToAPI(e *events.CloudEvent) *v1.CloudEvent {
	if e == nil {
		return nil
	}
	return &v1.CloudEvent{
		CloudEventsVersion: swag.String(e.CloudEventsVersion),
		ContentType:        e.ContentType,
		Data:               e.Data,
		EventID:            swag.String(e.EventID),
		EventTime:          strfmt.DateTime(e.EventTime),
		EventType:          swag.String(e.EventType),
		EventTypeVersion:   e.EventTypeVersion,
		Extensions:         e.Extensions,
		Namespace:          swag.String(e.Namespace),
		SchemaURL:          e.SchemaURL,
		SourceID:           swag.String(e.SourceID),
		SourceType:         swag.String(e.SourceType),
	}
}
