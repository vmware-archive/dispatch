///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package runtimes

import (
	"io"

	imagemanager "github.com/vmware/dispatch/pkg/image-manager"
)

// PowershellRuntime represents nodejs6 image suport
type PowershellRuntime struct {
	Language       imagemanager.Language
	PackageManager string
	ManifestFile   string
}

// Currently dependency management is not supported with powershell.  This
// requires more research into supporing NuGet or similar for managing packages
var powershellDockerfile = ``

// GetPackageManager returns the package manager
func (r *PowershellRuntime) GetPackageManager() string {
	return r.PackageManager
}

// PrepareManifest writes and adds the manifest to the Dockerfile
func (r *PowershellRuntime) PrepareManifest(dir string, image *imagemanager.Image) error {
	return nil
}

// WriteDockerfile writes the runtime Dockerfile
func (r *PowershellRuntime) WriteDockerfile(dockerfile io.Writer, image *imagemanager.Image) error {
	return nil
}

// NewPowershellRuntime returns a new runtime
func NewPowershellRuntime() *PowershellRuntime {
	return &PowershellRuntime{
		Language:       imagemanager.LanguagePowershell,
		PackageManager: "",
		ManifestFile:   "",
	}
}

func init() {
	imagemanager.RuntimeMap[imagemanager.LanguagePowershell] = NewPowershellRuntime()
}
