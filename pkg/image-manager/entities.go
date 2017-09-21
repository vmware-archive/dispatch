///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

// NO TESTS

import (
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

// const TypeBaseImage = "BaseImage"

const (
	// StatusINITIALIZED captures enum value "INITIALIZED"
	StatusINITIALIZED entitystore.Status = "INITIALIZED"
	// StatusCREATING captures enum value "CREATING"
	StatusCREATING entitystore.Status = "CREATING"
	// StatusREADY captures enum value "READY"
	StatusREADY entitystore.Status = "READY"
	// StatusERROR captures enum value "ERROR"
	StatusERROR entitystore.Status = "ERROR"
	// StatusDELETED captures enum value "DELETED"
	StatusDELETED entitystore.Status = "DELETED"
)

type BaseImage struct {
	entitystore.BaseEntity
	DockerURL string `json:"dockrUrl"`
	Public    bool   `json:"public"`
	SecretID  string `json:"secretId"`
}
