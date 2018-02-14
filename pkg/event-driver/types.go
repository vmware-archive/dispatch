///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventdriver

import (
	"github.com/vmware/dispatch/pkg/events"
)

// NO TESTS

// Driver understands 3rd party systems, consumes their events and publishes them to internal queue
type Driver interface {
	Run() error
	Close() error
}

// Consumer consumes events from 3rd party systems.
type Consumer interface {
	// Consume() takes a slice of topics to listen on, and returns a channel which generates events.
	// If len(topics) == 0, consumer should return all/default set of events.
	Consume(topics []string) (<-chan *events.CloudEvent, error)

	// Topics() returns list of available topics.
	Topics() []string

	// Should be called to stop consuming events.
	Close() error
}
