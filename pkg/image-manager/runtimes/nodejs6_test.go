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

func TestNodejs6Runtime(t *testing.T) {
	pkgJSON := `
{
	"name": "dispatch-ui",
	"version": "0.0.0",
	"license": "MIT",
	"scripts": {
		"ng": "ng",
		"start": "ng serve",
		"build": "ng build",
		"test": "ng test",
		"lint": "ng lint",
		"e2e": "ng e2e"
	},
	"private": true,
	"dependencies": {
		"@angular/animations": "^4.4.6",
		"@angular/common": "^4.4.6",
		"@angular/compiler": "^4.4.6",
		"@angular/core": "^4.4.6",
		"@angular/forms": "^4.4.6",
		"@angular/http": "^4.4.6",
		"@angular/platform-browser": "^4.4.6",
		"@angular/platform-browser-dynamic": "^4.4.6",
		"@angular/router": "^4.4.6",
		"@webcomponents/custom-elements": "^1.0.6",
		"clarity-angular": "^0.10.16",
		"clarity-icons": "^0.10.16",
		"clarity-ui": "^0.10.16",
		"core-js": "^2.5.1",
		"ng2-ace-editor": "^0.2.5",
		"prismjs": "^1.8.4",
		"rxjs": "^5.5.4",
		"zone.js": "^0.8.18"
	}
}`

	dockerfile := `
ADD package.json package.json
RUN npm install .
`
	i := &imagemanager.Image{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: imagemanager.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Language:  imagemanager.LanguagePython2,
		RuntimeDependencies: imagemanager.RuntimeDependencies{
			Manifest: pkgJSON,
		},
	}
	nodejs6Runtime := NewNodejs6Runtime()
	assert.Equal(t, "npm", nodejs6Runtime.GetPackageManager())
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	// Test manifest
	err = nodejs6Runtime.PrepareManifest(dir, i)
	assert.NoError(t, err)
	content, err := ioutil.ReadFile(fmt.Sprintf("%s/package.json", dir))
	assert.NoError(t, err)
	assert.Equal(t, i.RuntimeDependencies.Manifest, string(content))
	// Test Dockerfile
	b := new(bytes.Buffer)
	err = nodejs6Runtime.WriteDockerfile(b, i)
	assert.NoError(t, err)
	assert.Equal(t, dockerfile, b.String())
}
