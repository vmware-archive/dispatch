///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package applicationmanager

// NO TEST

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// Application is a data struct used to store application information into entity store
type Application struct {
	entitystore.BaseEntity
}
