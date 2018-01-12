///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

const (
	// FunctionSubscriber defines a function type of subscriber
	FunctionSubscriber = "function"
	// EventSubscriber defines an event type of subscriber
	EventSubscriber = "event"
)

// Subscription struct represents a single subscription of subscriber to publisher
type Subscription struct {
	entitystore.BaseEntity
	Topic      string     `json:"topic"`
	Subscriber Subscriber `json:"subscriber"`
	Secrets    []string   `json:"secrets,omitempty"`
}

// Subscriber represents a subscriber
type Subscriber struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// Driver represents a event driver, (e.g. vCenter)
type Driver struct {
	entitystore.BaseEntity
	Type    string            `json:"type"`
	Config  map[string]string `json:"config, omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
}
