///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package runtimes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	imagemanager "github.com/vmware/dispatch/pkg/image-manager"
)

func TestPython3Runtime(t *testing.T) {
	i := &imagemanager.Image{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePython3,
		RuntimeDependencies: imagemanager.RuntimeDependencies{
			Manifest: "python-dateutil==2.6.1\nPyYAML==3.12",
		},
	}
	python3Runtime := NewPython3Runtime()
	assert.Equal(t, "pip3", python3Runtime.GetPackageManager())
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	// Test manifest
	err = python3Runtime.PrepareManifest(dir, i)
	assert.NoError(t, err)
	content, err := ioutil.ReadFile(fmt.Sprintf("%s/requirements.txt", dir))
	assert.NoError(t, err)
	assert.Equal(t, i.RuntimeDependencies.Manifest, string(content))
	// Test Dockerfile
	dockerfile := new(bytes.Buffer)
	err = python3Runtime.WriteDockerfile(dockerfile, i)
	assert.NoError(t, err)
	assert.Equal(t, "\nADD requirements.txt requirements.txt\nRUN pip3 install -r requirements.txt\n", dockerfile.String())
}
