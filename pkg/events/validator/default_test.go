///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package validator_test

import (
	"encoding/json"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/validator"
)

var testEvent1 = events.CloudEvent{
	EventType:          "test.event",
	EventTypeVersion:   "0.1",
	CloudEventsVersion: events.CloudEventsVersion,
	Source:             "test.source.id",
	EventID:            uuid.NewV4().String(),
	EventTime:          time.Now(),
	SchemaURL:          "http://some.url.com/file",
	ContentType:        "application/json",
	Extensions:         nil,
	Data:               json.RawMessage(`{"example":"value"}`),
}

func TestDefaultValidate(t *testing.T) {
	v := validator.NewDefaultValidator()
	assert.NoError(t, v.Validate(&testEvent1))
	incorrect := testEvent1
	incorrect.EventID = ""
	assert.Error(t, v.Validate(&incorrect))
}
