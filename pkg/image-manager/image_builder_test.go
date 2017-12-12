///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

//go:generate mockery -name ImageAPIClient -case underscore -dir ../../vendor/github.com/docker/docker/client/

func mockBaseImageBuilder(es entitystore.EntityStore, client docker.ImageAPIClient) *BaseImageBuilder {

	return &BaseImageBuilder{
		baseImageChannel: make(chan BaseImage),
		done:             make(chan bool),
		es:               es,
		dockerClient:     client,
		orgID:            ImageManagerFlags.OrgID,
	}
}

func TestDockerDelete(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: StatusREADY,
		},
		DockerURL: "some/repo:latest",
		Public:    true,
	}

	es.Add(bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, nil).Once()
	assert.NoError(t, builder.dockerDelete(bi))

	es.Add(bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, fmt.Errorf("oh no")).Once()
	assert.Error(t, builder.dockerDelete(bi))

	bi.Status = StatusERROR
	es.Add(bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, fmt.Errorf("oh no")).Once()
	assert.NoError(t, builder.dockerDelete(bi))
}

func TestDockerStatus(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: StatusREADY,
		},
		DockerURL: "some/repo:latest",
		Public:    true,
	}

	es.Add(bi)
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{}, nil).Once()
	bis, err := builder.dockerStatus()
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
	assert.Equal(t, StatusERROR, bis[0].Status)

	summary := dockerTypes.ImageSummary{
		RepoTags: []string{bi.DockerURL},
	}
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{summary}, nil).Once()
	bis, err = builder.dockerStatus()
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
	assert.Equal(t, StatusREADY, bis[0].Status)

	es.Delete(builder.orgID, bi.Name, &BaseImage{})
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{summary}, nil).Once()
	bis, err = builder.dockerStatus()
	assert.NoError(t, err)
	assert.Len(t, bis, 0)
}

func TestDockerPull(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test",
			Status: StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Public:    true,
	}

	es.Add(bi)
	buffer := ioutil.NopCloser(bytes.NewBufferString(`{"error": "oops", "errorDetail": {"message": "oops detail"}}`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err := builder.dockerPull(bi)
	assert.NoError(t, err)
	assert.Equal(t, StatusERROR, bi.Status)

	buffer = ioutil.NopCloser(bytes.NewBufferString(`{"message": "yay"}`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err = builder.dockerPull(bi)
	assert.NoError(t, err)
	assert.Equal(t, StatusREADY, bi.Status)

	bi.Status = StatusINITIALIZED
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(nil, fmt.Errorf("bad image")).Once()
	err = builder.dockerPull(bi)
	assert.NoError(t, err)
	assert.Equal(t, StatusERROR, bi.Status)

	bi.Status = StatusINITIALIZED
	buffer = ioutil.NopCloser(bytes.NewBufferString(`{"message": "bad json"`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err = builder.dockerPull(bi)
	assert.Error(t, err)
	assert.Equal(t, StatusINITIALIZED, bi.Status)
}
