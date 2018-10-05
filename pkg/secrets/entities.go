///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package secrets

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// SecretEntity is the secret entity type
type SecretEntity struct {
	entitystore.BaseEntity
	Secrets map[string]string `json:"secrets"`
}
