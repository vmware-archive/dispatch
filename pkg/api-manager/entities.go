///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package apimanager

// NO TEST

import (
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// API is a data struct used to store api information into entity store
type API struct {
	entitystore.BaseEntity
	API gateway.API
}
