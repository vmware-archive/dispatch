///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import "github.com/vmware/dispatch/pkg/event-manager/drivers/entities"

// NO TEST

// Backend defines the event driver backend interface
type Backend interface {
	Deploy(*entities.Driver) error
	Update(*entities.Driver) error
	Delete(*entities.Driver) error
}
