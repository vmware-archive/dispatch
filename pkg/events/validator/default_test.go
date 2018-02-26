///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package validator_test

import (
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/validator"
)

var testEvent1 = events.CloudEvent{
	Namespace:          "dispatchframework.io",
	EventType:          "test.event",
	EventTypeVersion:   "0.1",
	CloudEventsVersion: events.CloudEventsVersion,
	SourceType:         "test.source",
	SourceID:           "test.source.id",
	EventID:            uuid.NewV4().String(),
	EventTime:          time.Now(),
	SchemaURL:          "http://some.url.com/file",
	ContentType:        "application/json",
	Extensions:         nil,
	Data:               []byte(`{"example":"value"}`),
}

func TestDefaultValidate(t *testing.T) {
	v := validator.NewDefaultValidator()
	assert.NoError(t, v.Validate(&testEvent1))
	incorrect := testEvent1
	incorrect.Namespace = ""
	assert.Error(t, v.Validate(&incorrect))
}
