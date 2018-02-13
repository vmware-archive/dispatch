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

// Python2Runtime represents nodejs6 image suport
type Python2Runtime struct {
	Language       imagemanager.Language
	PackageManager string
	ManifestFile   string
}

// GetPackageManager returns the package manager
func (r *Python2Runtime) GetPackageManager() string {
	return r.PackageManager
}

// PrepareManifest writes and adds the manifest to the Dockerfile
func (r *Python2Runtime) PrepareManifest(dir string, image *imagemanager.Image) error {
	manifestFileContent := []byte(image.RuntimeDependencies.Manifest)
	if err := ioutil.WriteFile(filepath.Join(dir, r.ManifestFile), manifestFileContent, 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", r.ManifestFile)
	}
	return nil
}

// WriteDockerfile writes the runtime Dockerfile
func (r *Python2Runtime) WriteDockerfile(dockerfile io.Writer, image *imagemanager.Image) error {
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

// NewPython2Runtime returns a new runtime
func NewPython2Runtime() *Python2Runtime {
	return &Python2Runtime{
		Language:       imagemanager.LanguagePython2,
		PackageManager: "pip",
		ManifestFile:   "requirements.txt",
	}
}

func init() {
	imagemanager.RuntimeMap[imagemanager.LanguagePython2] = NewPython2Runtime()
}
