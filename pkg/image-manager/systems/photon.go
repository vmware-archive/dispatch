///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package systems

import (
	"html/template"
	"io"

	"github.com/pkg/errors"
	imagemanager "github.com/vmware/dispatch/pkg/image-manager"
)

// PhotonSystem represents Photon system suport
type PhotonSystem struct {
	os             imagemanager.Os
	packageManager string
}

var photonDockerfile = `
FROM {{ .BaseImageURL }}
{{- if .Packages }}
RUN tdnf install -y \
{{- range .Packages }}
{{- if .Version }}
	{{ .Name }}-{{ .Version }} \
{{- else }}
	{{ .Name }} \
{{- end }}
{{- end }}
	&& tdnf clean all
{{- end }}
`

// GetPackageManager returns the systems pacakage manager
func (r *PhotonSystem) GetPackageManager() string {
	return r.packageManager
}

// WriteDockerfile writes out the dockerfile
func (r *PhotonSystem) WriteDockerfile(dockerfile io.Writer, baseImage *imagemanager.BaseImage, image *imagemanager.Image) error {
	tmpl, err := template.New(string(r.os)).Parse(photonDockerfile)
	if err != nil {
		return errors.Wrapf(err, "failed to build dockefile template")
	}
	args := struct {
		BaseImageURL string
		Packages     []imagemanager.SystemPackage
	}{
		BaseImageURL: baseImage.DockerURL,
		Packages:     image.SystemDependencies.Packages,
	}
	err = tmpl.Execute(dockerfile, args)
	if err != nil {
		return errors.Wrapf(err, "failed to write dockefile")
	}
	return nil
}

// NewPhotonSystem returns a new Photon system
func NewPhotonSystem() *PhotonSystem {
	return &PhotonSystem{
		os:             imagemanager.OsPhoton,
		packageManager: "tdnf",
	}
}

func init() {
	imagemanager.SystemMap[imagemanager.OsPhoton] = NewPhotonSystem()
}
