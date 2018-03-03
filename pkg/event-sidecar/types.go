///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventsidecar

// NO TESTS

// EventListener listens for new events, parses & validates them,
// and finally sends them using transport
type EventListener interface {
	Serve() error
	Shutdown() error
}
