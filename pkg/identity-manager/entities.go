///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

// NO TEST

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// Rule is a data struct to store rules within a policy
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

// ServiceAccount is a data struct used to store service accounts into entity store
type ServiceAccount struct {
	entitystore.BaseEntity
	PublicKey    string `json:"publicKey"`
	Domain       string `json:"domain"`
	JWTAlgorithm string `json:"jwtAlgorithm"`
}
