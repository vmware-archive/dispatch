///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package identitymanager

// NO TESTS

import (
	gooidc "github.com/coreos/go-oidc"
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

// Session is a Entity that stores a user session
type Session struct {
	entitystore.BaseEntity
	IDToken gooidc.IDToken
}
