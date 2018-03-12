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
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager"
)

func TestPowerShellRuntime(t *testing.T) {
	requirementsPSD := `@{
    psake        = 'latest'
    Pester       = 'latest'
    BuildHelpers = '0.0.20'
    PSDeploy     = '0.1.21'

    'RamblingCookieMonster/PowerShell' = 'master'
}
`

	dockerfile := `
ADD requirements.psd1 requirements.psd1
RUN pwsh -command '$ErrorActionPreference="Stop"; Invoke-PSDepend -Path "//requirements.psd1" -Confirm:$false'
`
	i := &imagemanager.Image{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePowershell,
		RuntimeDependencies: imagemanager.RuntimeDependencies{
			Manifest: requirementsPSD,
		},
	}
	powershellRuntime := NewPowershellRuntime()
	assert.Equal(t, "Invoke-PSDepend", powershellRuntime.GetPackageManager())
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	// Test manifest
	err = powershellRuntime.PrepareManifest(dir, i)
	assert.NoError(t, err)
	content, err := ioutil.ReadFile(fmt.Sprintf("%s/requirements.psd1", dir))
	assert.NoError(t, err)
	assert.Equal(t, i.RuntimeDependencies.Manifest, string(content))
	// Test Dockerfile
	b := new(bytes.Buffer)
	err = powershellRuntime.WriteDockerfile(b, i)
	assert.NoError(t, err)
	assert.Equal(t, dockerfile, b.String())
}
