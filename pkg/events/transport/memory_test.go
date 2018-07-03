///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/events"
)

func TestInMemoryPublish(t *testing.T) {
	event := events.NewCloudEventWithDefaults(testTopic)

	memory := NewInMemory()

	err := memory.Publish(context.Background(), &event, testTopic, testOrg)
	assert.NoError(t, err)
}

func TestInMemorySubscribe(t *testing.T) {
	event := events.NewCloudEventWithDefaults(testTopic)
	memory := NewInMemory()

	done := make(chan struct{})
	_, err := memory.Subscribe(context.Background(), testTopic, testOrg, func(ctx context.Context, e *events.CloudEvent) {
		assert.Equal(t, event.EventID, e.EventID)
		done <- struct{}{}
	})
	assert.NoError(t, err)

	err = memory.Publish(context.Background(), &event, testTopic, testOrg)
	assert.NoError(t, err)

	<-done

}
