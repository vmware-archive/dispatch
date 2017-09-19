///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

// NO TESTS

import (
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

const TypeBaseImage = "BaseImage"

type BaseImage struct {
	entitystore.BaseEntity
	DockerURL string `json:"dockrUrl"`
	Public    bool   `json:"public"`
	SecretID  string `json:"secretId"`
}
