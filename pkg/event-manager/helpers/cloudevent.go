///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package helpers

import (
	"time"

	"github.com/go-openapi/strfmt"
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
		EventType:          e.EventType,
		EventTypeVersion:   e.EventTypeVersion,
		CloudEventsVersion: e.CloudEventsVersion,
		Source:             e.Source,
		EventID:            e.EventID,
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
		CloudEventsVersion: e.CloudEventsVersion,
		ContentType:        e.ContentType,
		Data:               e.Data,
		EventID:            e.EventID,
		EventTime:          strfmt.DateTime(e.EventTime),
		EventType:          e.EventType,
		EventTypeVersion:   e.EventTypeVersion,
		Extensions:         e.Extensions,
		SchemaURL:          e.SchemaURL,
		Source:             e.Source,
	}
}
