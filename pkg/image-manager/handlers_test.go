///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

import (
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

func addEntity(t *testing.T, api *operations.ImageManagerAPI, h *Handlers, name, dockerURL string, public bool, tags map[string]string) *models.BaseImage {
	var entityTags []*models.Tag
	for k, v := range tags {
		entityTags = append(entityTags, &models.Tag{Key: k, Value: v})
	}

	reqBody := &models.BaseImage{
		Name:      swag.String(name),
		DockerURL: swag.String(dockerURL),
		Public:    swag.Bool(public),
		Tags:      entityTags,
	}
	r := httptest.NewRequest("POST", "/v1/image/base", nil)
	params := baseimage.AddBaseImageParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.BaseImageAddBaseImageHandler.Handle(params)
	var respBody models.BaseImage
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func TestBaseImageAddBaseImageHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	h := NewHandlers(nil)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addEntity(t, api, h, "testEntity", "test/base", true, map[string]string{"role": "test"})

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, "testEntity", *respBody.Name)
	assert.Equal(t, "test/base", *respBody.DockerURL)
	assert.Equal(t, true, *respBody.Public)
	assert.Equal(t, models.StatusINITIALIZED, respBody.Status)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestBaseImageGetBaseImageByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	h := NewHandlers(nil)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addEntity(t, api, h, "testEntity", "test/base", true, map[string]string{"role": "test"})

	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/image/base/testEntity", nil)
	get := baseimage.GetBaseImageByNameParams{
		HTTPRequest:   r,
		BaseImageName: "testEntity",
	}
	getResponder := api.BaseImageGetBaseImageByNameHandler.Handle(get)
	var getBody models.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "testEntity", *getBody.Name)
	assert.Equal(t, "test/base", *getBody.DockerURL)
	assert.Equal(t, true, *getBody.Public)
	assert.Equal(t, models.StatusINITIALIZED, getBody.Status)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)

	r = httptest.NewRequest("GET", "/v1/image/base/doesNotExist", nil)
	get = baseimage.GetBaseImageByNameParams{
		HTTPRequest:   r,
		BaseImageName: "doesNotExist",
	}
	getResponder = api.BaseImageGetBaseImageByNameHandler.Handle(get)
	helpers.HandlerRequest(t, getResponder, nil, 404)
}

func TestBaseImageGetBaseImagesHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	h := NewHandlers(nil)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addEntity(t, api, h, "testEntity1", "test/base", true, map[string]string{"role": "test", "item": "1"})
	addEntity(t, api, h, "testEntity2", "test/base", true, map[string]string{"role": "test", "item": "2"})
	addEntity(t, api, h, "testEntity3", "test/base", true, map[string]string{"role": "test", "item": "3"})

	r := httptest.NewRequest("GET", "/v1/image/base", nil)
	get := baseimage.GetBaseImagesParams{
		HTTPRequest: r,
	}
	getResponder := api.BaseImageGetBaseImagesHandler.Handle(get)
	var getBody []models.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 3)
}
