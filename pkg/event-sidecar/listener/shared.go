///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package listener

import "github.com/vmware/dispatch/pkg/events"

// SharedListener serves as a simple DI container for Listener.
type SharedListener struct {
	transport    events.Transport
	parser       events.StreamParser
	validator    events.Validator
	organization string
	driverType   string
}

// NewSharedListener creates new copy of SharedListener. It should not be used directly outside of the listener package.
func NewSharedListener(transport events.Transport, parser events.StreamParser, validator events.Validator, organization, driverType string) SharedListener {
	if transport == nil {
		panic("transport not set")
	}
	if validator == nil {
		panic("validator not set")
	}
	if parser == nil {
		panic("parser not set")
	}
	return SharedListener{
		transport:    transport,
		parser:       parser,
		validator:    validator,
		organization: organization,
		driverType:   driverType,
	}
}
