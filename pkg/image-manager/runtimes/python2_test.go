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

func TestPython2Runtime(t *testing.T) {
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
	python2Runtime := NewPython2Runtime()
	assert.Equal(t, "pip", python2Runtime.GetPackageManager())
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	// Test manifest
	err = python2Runtime.PrepareManifest(dir, i)
	assert.NoError(t, err)
	content, err := ioutil.ReadFile(fmt.Sprintf("%s/requirements.txt", dir))
	assert.NoError(t, err)
	assert.Equal(t, i.RuntimeDependencies.Manifest, string(content))
	// Test Dockerfile
	dockerfile := new(bytes.Buffer)
	err = python2Runtime.WriteDockerfile(dockerfile, i)
	assert.NoError(t, err)
	assert.Equal(t, "\nADD requirements.txt requirements.txt\nRUN pip install -r requirements.txt\n", dockerfile.String())
}
