///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/go-openapi/swag"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/images"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

//go:generate mockery -name ImageAPIClient -case underscore -dir ../../vendor/github.com/docker/docker/client/ -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

func init() {
	log.SetLevel(log.DebugLevel)
}

func mockBaseImageBuilder(es entitystore.EntityStore, client docker.ImageAPIClient) *BaseImageBuilder {

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		dockerClient:     client,
	}
}

func TestBaseImageDelete(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: StatusREADY,
		},
		DockerURL: "some/repo:latest",
	}

	es.Add(context.Background(), bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, nil).Once()
	assert.NoError(t, builder.baseImageDelete(context.Background(), bi))

	es.Add(context.Background(), bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, fmt.Errorf("oh no")).Once()
	assert.Error(t, builder.baseImageDelete(context.Background(), bi))

	bi.Status = StatusERROR
	es.Add(context.Background(), bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, fmt.Errorf("oh no")).Once()
	assert.NoError(t, builder.baseImageDelete(context.Background(), bi))
}

func TestBaseImageStatus(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi1 := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test-1",
			Status:         entitystore.StatusREADY,
			OrganizationID: "dispatch",
		},
		DockerURL: "some/repo:latest",
	}
	es.Add(context.Background(), bi1)

	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{}, nil).Once()
	bis, err := builder.baseImageStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
	assert.Equal(t, entitystore.StatusMISSING, bis[0].GetStatus())

	bi2 := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test-2",
			Status:         entitystore.StatusINITIALIZED,
			OrganizationID: "dispatch",
		},
		DockerURL: "some/repo:latest",
	}
	es.Add(context.Background(), bi2)

	summary := dockerTypes.ImageSummary{
		RepoTags: []string{bi1.DockerURL, bi2.DockerURL},
	}
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{summary}, nil).Once()
	bis, err = builder.baseImageStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
	assert.Equal(t, StatusREADY, bis[0].GetStatus())

	es.Delete(context.Background(), bi1.OrganizationID, bi1.Name, &BaseImage{})
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{summary}, nil).Once()
	bis, err = builder.baseImageStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
}

func TestBaseImagePull(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
	}

	es.Add(context.Background(), bi)
	buffer := ioutil.NopCloser(bytes.NewBufferString(`{"error": "oops", "errorDetail": {"message": "oops detail"}}`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err := builder.baseImagePull(context.Background(), bi)
	assert.Error(t, err)

	buffer = ioutil.NopCloser(bytes.NewBufferString(`{"message": "yay"}`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err = builder.baseImagePull(context.Background(), bi)
	assert.NoError(t, err)

	bi.Status = StatusINITIALIZED
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(nil, fmt.Errorf("bad image")).Once()
	err = builder.baseImagePull(context.Background(), bi)
	assert.Error(t, err)

	bi.Status = StatusINITIALIZED
	buffer = ioutil.NopCloser(bytes.NewBufferString(`{"message": "bad json"`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err = builder.baseImagePull(context.Background(), bi)
	assert.Error(t, err)
}

func Test_copyImageTemplate(t *testing.T) {
	dev.EnsureLocal(t)

	tmpDir, err := ioutil.TempDir("", "image-build")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	b, err := NewImageBuilder(nil, "", "")
	require.NoError(t, err)

	err = b.copyImageTemplate(tmpDir, "imikushin/dispatch-nodejs-base:0.0.2-dev1")
	require.NoError(t, err)

	bs, err := ioutil.ReadFile(filepath.Join(tmpDir, "Dockerfile"))
	require.NoError(t, err)
	assert.NotEmpty(t, bs)
}

func TestBuild(t *testing.T) {
	dev.EnsureLocal(t)

	tmpDir, err := ioutil.TempDir("", "image-build")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	b, err := NewImageBuilder(nil, "", "")
	require.NoError(t, err)

	nme := uuid.NewV4().String()
	image := &Image{
		DockerURL:           "imikushin/" + nme + ":latest",
		BaseImageName:       "js",
		SystemDependencies:  SystemDependencies{},
		RuntimeDependencies: RuntimeDependencies{},
	}

	err = b.copyImageTemplate(tmpDir, "dispatchframework/nodejs-base:0.0.7")
	require.NoError(t, err)

	spFile := filepath.Join(tmpDir, systemPackagesFile)
	require.NoError(t, b.writeSystemPackagesFile(spFile, image))

	pFile := filepath.Join(tmpDir, packagesFile)
	require.NoError(t, b.writePackagesFile(pFile, image))

	buildArgs := map[string]*string{
		"BASE_IMAGE":           swag.String("dispatchframework/nodejs-base:0.0.7"),
		"SYSTEM_PACKAGES_FILE": swag.String(systemPackagesFile),
		"PACKAGES_FILE":        swag.String(packagesFile),
	}

	err = images.Build(context.Background(), b.dockerClient, tmpDir, image.DockerURL, buildArgs)
	require.NoError(t, err)
}
