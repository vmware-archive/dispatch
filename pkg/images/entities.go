///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

// NO TESTS

import (
	"time"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

const (
	// StatusINITIALIZED captures enum value "INITIALIZED"
	StatusINITIALIZED entitystore.Status = "INITIALIZED"
	// StatusCREATING captures enum value "CREATING"
	StatusCREATING entitystore.Status = "CREATING"
	// StatusUPDATING capture enum value "UPDATING"
	StatusUPDATING entitystore.Status = "UPDATING"
	// StatusREADY captures enum value "READY"
	StatusREADY entitystore.Status = "READY"
	// StatusERROR captures enum value "ERROR"
	StatusERROR entitystore.Status = "ERROR"
	// StatusDELETED captures enum value "DELETED"
	StatusDELETED entitystore.Status = "DELETED"
)

// BaseImage defines a base image type
type BaseImage struct {
	entitystore.BaseEntity
	DockerURL    string    `json:"dockerUrl"`
	Language     string    `json:"language"`
	LastPullTime time.Time `json:"lastPullTime,omitempty"`
}

// SystemPackage defines a system package type
type SystemPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// SystemDependencies defines a system dependency type
type SystemDependencies struct {
	Packages []SystemPackage `json:"packages"`
}

// RuntimeDependencies defines a runtime dependency type
type RuntimeDependencies struct {
	Manifest string `json:"manifest"`
}

// Image defines an image type
type Image struct {
	entitystore.BaseEntity
	DockerURL           string              `json:"dockerUrl"`
	Language            string              `json:"language"`
	BaseImageName       string              `json:"baseImageName"`
	RuntimeDependencies RuntimeDependencies `json:"runtimeDependencies"`
	SystemDependencies  SystemDependencies  `json:"systemDependencies"`
	LastPullTime        time.Time           `json:"lastPullTime,omitempty"`
}

// GetDockerURL returns the docker URL for the image
func (i *Image) GetDockerURL() string {
	return i.DockerURL
}

// DockerImage defines the docker image interface
type DockerImage interface {
	entitystore.Entity
	GetDockerURL() string
}
