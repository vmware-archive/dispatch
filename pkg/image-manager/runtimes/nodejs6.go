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

// Nodejs6Runtime represents nodejs6 image suport
type Nodejs6Runtime struct {
	Language       imagemanager.Language
	PackageManager string
	ManifestFile   string
}

var nodejsDockerfile = `
ADD {{ .ManifestFile }} {{ .ManifestFile }}
RUN {{ .PackageManager }} install .
`

// GetPackageManager returns the package manager
func (r *Nodejs6Runtime) GetPackageManager() string {
	return r.PackageManager
}

// PrepareManifest writes and adds the manifest to the Dockerfile
func (r *Nodejs6Runtime) PrepareManifest(dir string, image *imagemanager.Image) error {
	if image.RuntimeDependencies.Manifest == "" {
		// skip if empty
		return nil
	}
	manifestFileContent := []byte(image.RuntimeDependencies.Manifest)
	if err := ioutil.WriteFile(filepath.Join(dir, r.ManifestFile), manifestFileContent, 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", r.ManifestFile)
	}
	return nil
}

// WriteDockerfile writes the runtime Dockerfile
func (r *Nodejs6Runtime) WriteDockerfile(dockerfile io.Writer, image *imagemanager.Image) error {
	if image.RuntimeDependencies.Manifest == "" {
		// skip if empty
		return nil
	}
	tmpl, err := template.New(string(r.Language)).Parse(nodejsDockerfile)
	if err != nil {
		return errors.Wrapf(err, "failed to build dockefile template")
	}
	err = tmpl.Execute(dockerfile, r)
	if err != nil {
		return errors.Wrapf(err, "failed to write dockefile")
	}
	return nil
}

// NewNodejs6Runtime returns a new runtime
func NewNodejs6Runtime() *Nodejs6Runtime {
	return &Nodejs6Runtime{
		Language:       imagemanager.LanguageNodejs6,
		PackageManager: "npm",
		ManifestFile:   "package.json",
	}
}

func init() {
	imagemanager.RuntimeMap[imagemanager.LanguageNodejs6] = NewNodejs6Runtime()
}
