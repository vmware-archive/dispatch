///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package events

// NO TESTS

// Event encapsulates an event that is transported via EventQueue
type Event struct {
	Topic       string
	ContentType string
	Body        []byte
	ID          string
}

// Handler is a callback function used to handle received event
type Handler func(event *Event)

// Subscription represents an active subscription within Queue. Subscription can be stopped
// by calling Unsubscribe()
type Subscription interface {
	Unsubscribe() error
}

// Queue is an abstraction over possible implementation of messaging
type Queue interface {
	Publish(event *Event) error
	Subscribe(topic string, handler Handler) (Subscription, error)
	Close()
}
