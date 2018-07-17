///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/vmware/dispatch/pkg/testing/dev"
	"github.com/vmware/dispatch/pkg/utils"
)

func TestImageName(t *testing.T) {
	prefix := rand.String(9)
	fnID := rand.String(6)
	assert.Equal(t, prefix+"/func-"+fnID+":latest", imageName(prefix, fnID))
}

func tarGzBytes(t *testing.T) []byte {
	tmpDir, err := ioutil.TempDir("", "tmp-src-dir")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.MkdirAll(filepath.Join(tmpDir, "mypkg"), 0775)
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tmpDir, "mypkg", "myfunc.py"), []byte("def hello(): pass"), 0664)
	require.NoError(t, err)

	bs, err := utils.TarGzBytes(tmpDir)
	require.NoError(t, err)

	return bs
}

func TestWriteSourceDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "func-build")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	err = writeSourceDir(tmpDir, tarGzBytes(t))
	require.NoError(t, err)

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "mypkg", "myfunc.py"))
	require.NoError(t, err)
	assert.Equal(t, "def hello(): pass", string(b))

	ls, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 1, len(ls))
	assert.Equal(t, "mypkg", ls[0].Name())
}

func Test_copyFunctionTemplate(t *testing.T) {
	dev.EnsureLocal(t)

	tmpDir, err := ioutil.TempDir("", "image-build")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dc, err := docker.NewEnvClient()
	require.NoError(t, err)
	b := NewDockerImageBuilder("", "", dc)
	require.NoError(t, err)

	err = b.copyFunctionTemplate(tmpDir, "imikushin/dispatch-nodejs-base:0.0.2-dev1")
	require.NoError(t, err)

	bs, err := ioutil.ReadFile(filepath.Join(tmpDir, "Dockerfile"))
	require.NoError(t, err)
	assert.NotEmpty(t, bs)
}
