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
	// StatusUPDATING capture enum value "UPDATING"
	StatusUPDATING entitystore.Status = "UPDATING"
	// StatusREADY captures enum value "READY"
	StatusREADY entitystore.Status = "READY"
	// StatusERROR captures enum value "ERROR"
	StatusERROR entitystore.Status = "ERROR"
	// StatusDELETED captures enum value "DELETED"
	StatusDELETED entitystore.Status = "DELETED"
)

// Language specification type
type Language string

const (
	// LanguagePython2 captures enum value "python2"
	LanguagePython2 Language = "python2"
	// LanguagePython3 captures enum value "python3"
	LanguagePython3 Language = "python3"
	// LanguageNodejs6 captures enum value "nodejs6"
	LanguageNodejs6 Language = "nodejs6"
	// LanguagePowershell captures enum value "powershell"
	LanguagePowershell Language = "powershell"
)

// Os specification type
type Os string

const (
	// OsPhoton captures enum value "photon"
	OsPhoton Os = "photon"
)

// BaseImage defines a base image type
type BaseImage struct {
	entitystore.BaseEntity
	DockerURL string   `json:"dockerUrl"`
	Language  Language `json:"language"`
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
	Format   string `json:"format"`
	Manifest string `json:"manifest"`
}

// Image defines an image type
type Image struct {
	entitystore.BaseEntity
	DockerURL           string              `json:"dockerUrl"`
	Language            Language            `json:"language"`
	BaseImageName       string              `json:"baseImageName"`
	RuntimeDependencies RuntimeDependencies `json:"runtimeDependencies"`
	SystemDependencies  SystemDependencies  `json:"systemDependencies"`
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
