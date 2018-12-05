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
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/go-openapi/swag"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/images"

	"github.com/vmware/dispatch/pkg/entity-store"
	mocks "github.com/vmware/dispatch/pkg/mocks/docker"
	testutils "github.com/vmware/dispatch/pkg/testing"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

//go:generate mockery -name CommonAPIClient -case underscore -dir ../../vendor/github.com/docker/docker/client/ -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

const testPullPeriod = time.Duration(5 * time.Minute)

func init() {
	log.SetLevel(log.DebugLevel)
}

func mockBaseImageBuilder(es entitystore.EntityStore, client docker.CommonAPIClient) *BaseImageBuilder {

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		dockerClient:     client,
		pullPeriod:       testPullPeriod,
	}
}

func mockImageBuilder(es entitystore.EntityStore, client docker.CommonAPIClient) *ImageBuilder {

	return &ImageBuilder{
		imageChannel: make(chan Image),
		done:         make(chan bool),
		es:           es,
		dockerClient: client,
		pullPeriod:   testPullPeriod,
	}
}

func TestBaseImageDelete(t *testing.T) {
	client := &mocks.CommonAPIClient{}
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
	client.On("ImageRemove", mock.Anything, bi.DockerURL, mock.Anything).Return(nil, nil).Once()
	assert.NoError(t, builder.baseImageDelete(context.Background(), bi))

	es.Add(context.Background(), bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, mock.Anything).Return(nil, fmt.Errorf("oh no")).Once()
	assert.Error(t, builder.baseImageDelete(context.Background(), bi))

	bi.Status = StatusERROR
	es.Add(context.Background(), bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, mock.Anything).Return(nil, fmt.Errorf("oh no")).Once()
	assert.NoError(t, builder.baseImageDelete(context.Background(), bi))
}

func TestBaseImageStatus(t *testing.T) {
	client := &mocks.CommonAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi1 := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test-1",
			Status:         entitystore.StatusREADY,
			OrganizationID: "dispatch",
		},
		DockerURL:    "some/repo:latest",
		LastPullTime: time.Now().Add(-testPullPeriod),
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

func TestBaseImageStatusPullPeriod(t *testing.T) {
	client := &mocks.CommonAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi1 := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test-1",
			Status:         entitystore.StatusREADY,
			OrganizationID: "dispatch",
		},
		DockerURL:    "some/repo:latest",
		LastPullTime: time.Now(),
	}
	es.Add(context.Background(), bi1)

	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{}, nil).Once()
	bis, err := builder.baseImageStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, bis, 0)
}

func TestBaseImagePull(t *testing.T) {
	client := &mocks.CommonAPIClient{}
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

	tmpDir, err := ioutil.TempDir("", "image-build")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	b, err := NewImageBuilder(nil, "", "", testPullPeriod)
	require.NoError(t, err)

	client := &mocks.CommonAPIClient{}
	b.dockerClient = client

	client.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.ContainerCreateCreatedBody{}, nil)
	client.On("ContainerRemove", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	client.On("ContainerInspect", mock.Anything, mock.Anything).Return(
		dockerTypes.ContainerJSON{Config: &container.Config{Labels: make(map[string]string)}}, nil)

	archive := testutils.TarArchive([]testutils.TestFile{{"Dockerfile", "This is a  dockerfile."}})
	client.On("CopyFromContainer", mock.Anything, mock.Anything, mock.Anything).Return(archive, dockerTypes.ContainerPathStat{}, nil)

	err = b.copyImageTemplate(tmpDir, "imikushin/dispatch-nodejs-base:0.0.2-dev1")
	require.NoError(t, err)

	bs, err := ioutil.ReadFile(filepath.Join(tmpDir, "Dockerfile"))
	require.NoError(t, err)
	assert.NotEmpty(t, bs)
}

func TestBuild(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "image-build")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	b, err := NewImageBuilder(nil, "", "", testPullPeriod)
	require.NoError(t, err)

	client := &mocks.CommonAPIClient{}
	b.dockerClient = client

	client.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.ContainerCreateCreatedBody{}, nil)
	client.On("ContainerRemove", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	client.On("ContainerInspect", mock.Anything, mock.Anything).Return(
		dockerTypes.ContainerJSON{Config: &container.Config{Labels: make(map[string]string)}}, nil)

	archive := testutils.TarArchive([]testutils.TestFile{{"Dockerfile", "This is a  dockerfile."}})
	client.On("CopyFromContainer", mock.Anything, mock.Anything, mock.Anything).Return(archive, dockerTypes.ContainerPathStat{}, nil)
	client.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).Return(dockerTypes.ImageBuildResponse{
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"message":"test"}`)),
	}, nil)
	nme := uuid.NewV4().String()
	image := &Image{
		DockerURL:           "imikushin/" + nme + ":latest",
		BaseImageName:       "js",
		SystemDependencies:  SystemDependencies{},
		RuntimeDependencies: RuntimeDependencies{},
	}

	err = b.copyImageTemplate(tmpDir, "dispatchframework/nodejs-base:0.0.8")
	require.NoError(t, err)

	spFile := filepath.Join(tmpDir, systemPackagesFile)
	require.NoError(t, b.writeSystemPackagesFile(spFile, image))

	pFile := filepath.Join(tmpDir, packagesFile)
	require.NoError(t, b.writePackagesFile(pFile, image))

	buildArgs := map[string]*string{
		"BASE_IMAGE":           swag.String("dispatchframework/nodejs-base:0.0.8"),
		"SYSTEM_PACKAGES_FILE": swag.String(systemPackagesFile),
		"PACKAGES_FILE":        swag.String(packagesFile),
	}

	err = images.Build(context.Background(), b.dockerClient, tmpDir, image.DockerURL, buildArgs)

	require.NoError(t, err)
}

func TestImageStatusPullPeriod(t *testing.T) {
	client := &mocks.CommonAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockImageBuilder(es, client)

	bi1 := &Image{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test-1",
			Status:         entitystore.StatusREADY,
			OrganizationID: "dispatch",
		},
		DockerURL:    "some/repo:latest",
		LastPullTime: time.Now(),
	}
	es.Add(context.Background(), bi1)

	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{}, nil).Once()
	images, err := builder.imageStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, images, 0)

	bi2 := &Image{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test-2",
			Status:         entitystore.StatusREADY,
			OrganizationID: "dispatch",
		},
		DockerURL:    "some/repo:latest",
		LastPullTime: time.Now().Add(-testPullPeriod),
	}
	es.Add(context.Background(), bi2)

	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{}, nil).Once()
	images, err = builder.imageStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, images, 1)
	assert.Equal(t, entitystore.StatusMISSING, images[0].GetStatus())
}
