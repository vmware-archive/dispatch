///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package listener

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/mocks"
)

func mockSharedListener() SharedListener {
	return SharedListener{
		transport: &mocks.Transport{},
		parser:    &mocks.StreamParser{},
		validator: &mocks.Validator{},
	}
}

func emptySharedListener() SharedListener {
	return SharedListener{}
}

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
	Data:               `{"example":"value"}`,
}

func eventJSON(event *events.CloudEvent) []byte {
	val, _ := json.Marshal(event)
	return val
}

func TestNewSharedListener(t *testing.T) {
	assert.Panics(t, func() { NewSharedListener(nil, nil, nil, "", "") })
	assert.Panics(t, func() { NewSharedListener(&mocks.Transport{}, nil, nil, "", "") })
	assert.Panics(t, func() { NewSharedListener(&mocks.Transport{}, &mocks.StreamParser{}, nil, "", "") })
	assert.NotPanics(t, func() { NewSharedListener(&mocks.Transport{}, &mocks.StreamParser{}, &mocks.Validator{}, "", "") })
}
