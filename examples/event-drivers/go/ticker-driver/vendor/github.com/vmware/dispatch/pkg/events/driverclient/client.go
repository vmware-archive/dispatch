///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package driverclient

import "github.com/vmware/dispatch/pkg/events"

// Client specifies an interface that event-driver clients must implement.
type Client interface {
	Send(event []events.CloudEvent) error
	SendOne(event *events.CloudEvent) error
	Validate(event []events.CloudEvent) error
	ValidateOne(event *events.CloudEvent) error
}
