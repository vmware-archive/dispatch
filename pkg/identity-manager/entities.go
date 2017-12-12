///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package identitymanager

// NO TESTS

import (
	gooidc "github.com/coreos/go-oidc"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// Session is a Entity that stores a user session
type Session struct {
	entitystore.BaseEntity
	IDToken gooidc.IDToken
}
