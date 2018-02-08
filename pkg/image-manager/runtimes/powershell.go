///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package runtimes

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	imagemanager "github.com/vmware/dispatch/pkg/image-manager"
)

// PowershellRuntime represents nodejs6 image suport
type PowershellRuntime struct {
	Language       imagemanager.Language
	PackageManager string
	ManifestFile   string
}

var powershellDockerfile = `
ADD {{ .ManifestFile }} {{ .ManifestFile }}
RUN {{ .PackageManager }} restore .
`

// GetPackageManager returns the package manager
func (r *PowershellRuntime) GetPackageManager() string {
	return r.PackageManager
}

// PrepareManifest writes and adds the manifest to the Dockerfile
func (r *PowershellRuntime) PrepareManifest(dir string, image *imagemanager.Image) error {
	manifestFileContent := []byte(image.RuntimeDependencies.Manifest)
	if err := ioutil.WriteFile(filepath.Join(dir, r.ManifestFile), manifestFileContent, 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", r.ManifestFile)
	}
	return nil
}

// WriteDockerfile writes the runtime Dockerfile
func (r *PowershellRuntime) WriteDockerfile(dockerfile io.Writer, image *imagemanager.Image) error {
	tmpl, err := template.New(string(r.Language)).Parse(powershellDockerfile)
	if err != nil {
		return errors.Wrapf(err, "failed to build dockefile template")
	}
	err = tmpl.Execute(dockerfile, r)
	if err != nil {
		return errors.Wrapf(err, "failed to write dockefile")
	}
	return nil
}

// NewPowershellRuntime returns a new runtime
func NewPowershellRuntime() *PowershellRuntime {
	return &PowershellRuntime{
		Language:       imagemanager.LanguagePowershell,
		PackageManager: "nuget",
		ManifestFile:   "packages.config",
	}
}

func init() {
	imagemanager.RuntimeMap[imagemanager.LanguagePowershell] = NewPowershellRuntime()
}
