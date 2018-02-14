///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// Subscription struct represents a single subscription of subscriber to publisher
type Subscription struct {
	entitystore.BaseEntity
	EventType  string   `json:"eventType"`
	SourceType string   `json:"sourceType"`
	SourceName string   `json:"sourceName"`
	Function   string   `json:"function"`
	Secrets    []string `json:"secrets,omitempty"`
}

// Driver represents an event driver, (e.g. vCenter)
type Driver struct {
	entitystore.BaseEntity
	Type    string            `json:"type"`
	Config  map[string]string `json:"config, omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Image   string            `json:"image"`
	Mode    string            `josn:"mode"`
}

// DriverType represents a custom type of driver
type DriverType struct {
	entitystore.BaseEntity
	Image  string            `json:"image"`
	Mode   string            `json:"mode"`
	Config map[string]string `json:"config,omitempty"`
}
