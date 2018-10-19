///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package driverclient

import "github.com/vmware/dispatch/pkg/events"

// PipeClient implements Event Driver client using named pipes.
type PipeClient struct {
}

// Send sends slice of events to event driver sidecar
func (PipeClient) Send(event []events.CloudEvent) error {
	panic("implement me")
}

// SendOne sends single event to event driver sidecar
func (PipeClient) SendOne(event *events.CloudEvent) error {
	panic("implement me")
}

// Validate validates list of events without sending them.
func (PipeClient) Validate(event []events.CloudEvent) error {
	panic("implement me")
}

// ValidateOne validates single event without sending it.
func (PipeClient) ValidateOne(event *events.CloudEvent) error {
	panic("implement me")
}
