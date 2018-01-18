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
		Public:    true,
	}

	es.Add(bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, nil).Once()
	assert.NoError(t, builder.baseImageDelete(bi))

	es.Add(bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, fmt.Errorf("oh no")).Once()
	assert.Error(t, builder.baseImageDelete(bi))

	bi.Status = StatusERROR
	es.Add(bi)
	client.On("ImageRemove", mock.Anything, bi.DockerURL, dockerTypes.ImageRemoveOptions{Force: true}).Return(nil, fmt.Errorf("oh no")).Once()
	assert.NoError(t, builder.baseImageDelete(bi))
}

func TestBaseImageStatus(t *testing.T) {
	client := &mocks.ImageAPIClient{}
	es := helpers.MakeEntityStore(t)
	builder := mockBaseImageBuilder(es, client)

	bi1 := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test-1",
			Status: entitystore.StatusREADY,
		},
		DockerURL: "some/repo:latest",
		Public:    true,
	}
	es.Add(bi1)

	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{}, nil).Once()
	bis, err := builder.baseImageStatus()
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
	assert.Equal(t, entitystore.StatusMISSING, bis[0].GetStatus())

	bi2 := &BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   "test-2",
			Status: entitystore.StatusINITIALIZED,
		},
		DockerURL: "some/repo:latest",
		Public:    true,
	}
	es.Add(bi2)

	summary := dockerTypes.ImageSummary{
		RepoTags: []string{bi1.DockerURL, bi2.DockerURL},
	}
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{summary}, nil).Once()
	bis, err = builder.baseImageStatus()
	assert.NoError(t, err)
	assert.Len(t, bis, 1)
	assert.Equal(t, StatusREADY, bis[0].GetStatus())

	es.Delete(builder.orgID, bi1.Name, &BaseImage{})
	client.On("ImageList", mock.Anything, dockerTypes.ImageListOptions{All: false}).Return([]dockerTypes.ImageSummary{summary}, nil).Once()
	bis, err = builder.baseImageStatus()
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
		Public:    true,
	}

	es.Add(bi)
	buffer := ioutil.NopCloser(bytes.NewBufferString(`{"error": "oops", "errorDetail": {"message": "oops detail"}}`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err := builder.baseImagePull(bi)
	assert.Error(t, err)

	buffer = ioutil.NopCloser(bytes.NewBufferString(`{"message": "yay"}`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err = builder.baseImagePull(bi)
	assert.NoError(t, err)

	bi.Status = StatusINITIALIZED
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(nil, fmt.Errorf("bad image")).Once()
	err = builder.baseImagePull(bi)
	assert.Error(t, err)

	bi.Status = StatusINITIALIZED
	buffer = ioutil.NopCloser(bytes.NewBufferString(`{"message": "bad json"`))
	client.On("ImagePull", mock.Anything, bi.DockerURL, dockerTypes.ImagePullOptions{All: false}).Return(buffer, nil).Once()
	err = builder.baseImagePull(bi)
	assert.Error(t, err)
}
