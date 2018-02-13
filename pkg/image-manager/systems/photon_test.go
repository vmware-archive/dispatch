///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package systems

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	imagemanager "github.com/vmware/dispatch/pkg/image-manager"
)

func TestPhotonSystemWithDeps(t *testing.T) {
	dockerfile := `
FROM some/repo:latest
RUN tdnf install -y \
	pkg1 \
	pkg2-0.1.1 \
	&& tdnf clean all
`
	bi := &imagemanager.BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePython3,
	}
	i := &imagemanager.Image{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePython2,
		SystemDependencies: imagemanager.SystemDependencies{
			Packages: []imagemanager.SystemPackage{
				imagemanager.SystemPackage{
					Name: "pkg1",
				},
				imagemanager.SystemPackage{
					Name:    "pkg2",
					Version: "0.1.1",
				},
			},
		},
		RuntimeDependencies: imagemanager.RuntimeDependencies{
			Manifest: "python-dateutil==2.6.1\nPyYAML==3.12",
		},
	}
	photonSystem := NewPhotonSystem()
	assert.Equal(t, "tdnf", photonSystem.GetPackageManager())

	b := new(bytes.Buffer)
	err := photonSystem.WriteDockerfile(b, bi, i)
	assert.NoError(t, err)
	assert.Equal(t, dockerfile, b.String())
}

func TestPhotonSystemNoDeps(t *testing.T) {
	dockerfile := `
FROM some/repo:latest
`
	bi := &imagemanager.BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePython3,
	}
	i := &imagemanager.Image{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePython2,
		RuntimeDependencies: imagemanager.RuntimeDependencies{
			Manifest: "python-dateutil==2.6.1\nPyYAML==3.12",
		},
	}
	photonSystem := NewPhotonSystem()
	assert.Equal(t, "tdnf", photonSystem.GetPackageManager())

	b := new(bytes.Buffer)
	err := photonSystem.WriteDockerfile(b, bi, i)
	assert.NoError(t, err)
	assert.Equal(t, dockerfile, b.String())
}
