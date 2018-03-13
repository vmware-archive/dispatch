///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package runtimes

import (
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/image-manager"
)

// PowershellRuntime represents nodejs6 image suport
type PowershellRuntime struct {
	Language       imagemanager.Language
	PackageManager string
	ManifestFile   string
}

// Uses PSDepend as dependency manager: https://github.com/RamblingCookieMonster/PSDepend
var powershellDockerfile = `
ADD {{ .ManifestFile }} {{ .ManifestFile }}
RUN pwsh -command '$ErrorActionPreference="Stop"; {{ .PackageManager }} -Path "//{{ .ManifestFile }}" -Confirm:$false'
`

// GetPackageManager returns the package manager
func (r *PowershellRuntime) GetPackageManager() string {
	return r.PackageManager
}

// PrepareManifest writes and adds the manifest to the Dockerfile
func (r *PowershellRuntime) PrepareManifest(dir string, image *imagemanager.Image) error {
	manifestFileContent := []byte(image.RuntimeDependencies.Manifest)
	if len(manifestFileContent) == 0 {
		// Don't create manifest, PSDepend does not work with empty manifest
		return nil
	}
	if err := ioutil.WriteFile(filepath.Join(dir, r.ManifestFile), manifestFileContent, 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", r.ManifestFile)
	}
	return nil
}

// WriteDockerfile writes the runtime Dockerfile
func (r *PowershellRuntime) WriteDockerfile(dockerfile io.Writer, image *imagemanager.Image) error {
	if len(image.RuntimeDependencies.Manifest) == 0 {
		// PSDepend does not work with empty manifest
		return nil
	}
	tmpl, err := template.New(string(r.Language)).Parse(powershellDockerfile)
	if err != nil {
		return errors.Wrapf(err, "failed to build dockerfile template")
	}
	err = tmpl.Execute(dockerfile, r)
	if err != nil {
		return errors.Wrapf(err, "failed to write dockerfile")
	}
	return nil
}

// NewPowershellRuntime returns a new runtime
func NewPowershellRuntime() *PowershellRuntime {
	return &PowershellRuntime{
		Language:       imagemanager.LanguagePowershell,
		PackageManager: "Invoke-PSDepend",
		ManifestFile:   "requirements.psd1",
	}
}

func init() {
	imagemanager.RuntimeMap[imagemanager.LanguagePowershell] = NewPowershellRuntime()
}
