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

// Python3Runtime represents nodejs6 image suport
type Python3Runtime struct {
	Language       imagemanager.Language
	PackageManager string
	ManifestFile   string
}

var pythonDockerfile = `
ADD {{ .ManifestFile }} {{ .ManifestFile }}
RUN {{ .PackageManager }} install -r {{ .ManifestFile }}
`

// GetPackageManager returns the package manager
func (r *Python3Runtime) GetPackageManager() string {
	return r.PackageManager
}

// PrepareManifest writes and adds the manifest to the Dockerfile
func (r *Python3Runtime) PrepareManifest(dir string, image *imagemanager.Image) error {
	manifestFileContent := []byte(image.RuntimeDependencies.Manifest)
	if err := ioutil.WriteFile(filepath.Join(dir, r.ManifestFile), manifestFileContent, 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", r.ManifestFile)
	}
	return nil
}

// WriteDockerfile writes the runtime Dockerfile
func (r *Python3Runtime) WriteDockerfile(dockerfile io.Writer, image *imagemanager.Image) error {
	tmpl, err := template.New(string(r.Language)).Parse(pythonDockerfile)
	if err != nil {
		return errors.Wrapf(err, "failed to build dockefile template")
	}
	err = tmpl.Execute(dockerfile, r)
	if err != nil {
		return errors.Wrapf(err, "failed to write dockefile")
	}
	return nil
}

// NewPython3Runtime returns a new runtime
func NewPython3Runtime() *Python3Runtime {
	return &Python3Runtime{
		Language:       imagemanager.LanguagePython3,
		PackageManager: "pip3",
		ManifestFile:   "requirements.txt",
	}
}

func init() {
	imagemanager.RuntimeMap[imagemanager.LanguagePython3] = NewPython3Runtime()
}
