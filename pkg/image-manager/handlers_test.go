///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations"
	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations/base_image"
	helpers "gitlab.eng.vmware.com/serverless/serverless/pkg/testing/api"
)

type testEntity struct {
	entitystore.BaseEntity
	Value string `json:"value"`
}

func TestBaseImageAddBaseImageHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	h := NewImageManagerHandlers(nil)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	var tags []*models.Tag
	tags = append(tags, &models.Tag{Key: "role", Value: "test"})
	reqBody := &models.BaseImage{
		Name:      swag.String("testEntity"),
		DockerURL: swag.String("test/base"),
		Public:    swag.Bool(true),
		Tags:      tags,
	}
	r := httptest.NewRequest("POST", "/v1/image/base", nil)
	params := baseimage.AddBaseImageParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.BaseImageAddBaseImageHandler.Handle(params)
	var respBody models.BaseImage
	helpers.HandlerRequest(t, responder, &respBody, 201)

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, reqBody.Name, respBody.Name)
	assert.Equal(t, reqBody.DockerURL, respBody.DockerURL)
	assert.Equal(t, reqBody.Public, respBody.Public)
	assert.Equal(t, models.StatusINITIALIZED, respBody.Status)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestBaseImageGetBaseImageByIDHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	h := NewImageManagerHandlers(nil)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	var tags []*models.Tag
	tags = append(tags, &models.Tag{Key: "role", Value: "test"})
	reqBody := &models.BaseImage{
		Name:      swag.String("testEntity"),
		DockerURL: swag.String("test/base"),
		Public:    swag.Bool(true),
		Tags:      tags,
	}
	r := httptest.NewRequest("POST", "/v1/image/base", nil)
	add := baseimage.AddBaseImageParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	addResponder := api.BaseImageAddBaseImageHandler.Handle(add)
	var addBody models.BaseImage
	helpers.HandlerRequest(t, addResponder, &addBody, 201)

	assert.NotEmpty(t, addBody.ID)

	id := addBody.ID
	createdTime := addBody.CreatedTime
	r = httptest.NewRequest("GET", fmt.Sprintf("/v1/image/base/%v", id), nil)
	get := baseimage.GetBaseImageByIDParams{
		HTTPRequest: r,
		BaseImageID: id,
	}
	getResponder := api.BaseImageGetBaseImageByIDHandler.Handle(get)
	var getBody models.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, id, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, reqBody.Name, getBody.Name)
	assert.Equal(t, reqBody.DockerURL, getBody.DockerURL)
	assert.Equal(t, reqBody.Public, getBody.Public)
	assert.Equal(t, models.StatusINITIALIZED, getBody.Status)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)
}
