///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

// NO TESTS

import (
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

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
	DockerURL string `json:"dockerUrl"`
	Language  string `json:"language"`
	Public    bool   `json:"public"`
}

type Image struct {
	entitystore.BaseEntity
	DockerURL     string `json:"dockerUrl"`
	Language      string `json:"language"`
	BaseImageName string `json:"baseImageName"`
}

func (i *Image) GetDockerURL() string {
	return i.DockerURL
}

type DockerImage interface {
	entitystore.Entity
	GetDockerURL() string
}
