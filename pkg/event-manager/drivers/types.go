///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"

	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
)

// NO TEST

// Backend defines the event driver backend interface
type Backend interface {
	Deploy(context.Context, *entities.Driver) error
	Update(context.Context, *entities.Driver) error
	Delete(context.Context, *entities.Driver) error
}
