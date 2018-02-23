///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

// NO TEST

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

type Rule struct {
	entitystore.BaseEntity
	Subjects  []string `json:"subjects"`
	Resources []string `json:"resources"`
	Actions   []string `json:"actions"`
}

// Policy is a data struct used to store policy into entity store
type Policy struct {
	entitystore.BaseEntity
	Rules []Rule `json:"rules"`
}
