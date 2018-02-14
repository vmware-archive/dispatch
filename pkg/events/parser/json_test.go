///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package parser

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/events"
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

func eventJSON(event *events.CloudEvent) []byte {
	val, _ := json.Marshal(event)
	return val
}

func TestParsingEmptyList(t *testing.T) {
	buf := bytes.NewBufferString("[]")

	p := &JSONEventParser{}
	evs, err := p.Parse(buf)
	assert.NoError(t, err)
	assert.Len(t, evs, 0)
}

func TestParsingMalformed(t *testing.T) {
	buf := bytes.NewBufferString("{gdsgsdgs}")

	p := &JSONEventParser{}
	evs, err := p.Parse(buf)
	assert.Error(t, err)
	assert.Len(t, evs, 0)
}

func TestParsingCorrect(t *testing.T) {
	buf := bytes.NewBuffer(eventJSON(&testEvent1))
	p := &JSONEventParser{}
	evs, err := p.Parse(buf)
	assert.NoError(t, err)
	assert.Len(t, evs, 1)
	assert.True(t, cmp.Equal(testEvent1, evs[0]))
}
