///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations"
	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations/base_image"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations/image"
	helpers "gitlab.eng.vmware.com/serverless/serverless/pkg/testing/api"
)

type testEntity struct {
	entitystore.BaseEntity
	Value string `json:"value"`
}

func addBaseImageEntity(t *testing.T, api *operations.ImageManagerAPI, h *Handlers, name, dockerURL string, public bool, tags map[string]string) *models.BaseImage {
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
	responder := api.BaseImageAddBaseImageHandler.Handle(params, "testCookie")
	var respBody models.BaseImage
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func addImageEntity(t *testing.T, api *operations.ImageManagerAPI, h *Handlers, name, baseImageName string, tags map[string]string) *models.Image {
	var entityTags []*models.Tag
	for k, v := range tags {
		entityTags = append(entityTags, &models.Tag{Key: k, Value: v})
	}

	reqBody := &models.Image{
		Name:          swag.String(name),
		BaseImageName: swag.String(baseImageName),
		Tags:          entityTags,
	}
	r := httptest.NewRequest("POST", "/v1/image", nil)
	params := image.AddImageParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.ImageAddImageHandler.Handle(params, "testCookie")
	var respBody models.Image
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func TestBaseImageAddBaseImageHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addBaseImageEntity(t, api, h, "testEntity", "test/base", true, map[string]string{"role": "test"})

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
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addBaseImageEntity(t, api, h, "testEntity", "test/base", true, map[string]string{"role": "test"})

	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/image/base/testEntity", nil)
	get := baseimage.GetBaseImageByNameParams{
		HTTPRequest:   r,
		BaseImageName: "testEntity",
	}
	getResponder := api.BaseImageGetBaseImageByNameHandler.Handle(get, "testCookie")
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
	getResponder = api.BaseImageGetBaseImageByNameHandler.Handle(get, "testCookie")

	var errorBody models.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestBaseImageGetBaseImagesHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testEntity1", "test/base", true, map[string]string{"role": "test", "item": "1"})
	addBaseImageEntity(t, api, h, "testEntity2", "test/base", true, map[string]string{"role": "test", "item": "2"})
	addBaseImageEntity(t, api, h, "testEntity3", "test/base", true, map[string]string{"role": "test", "item": "3"})

	r := httptest.NewRequest("GET", "/v1/image/base", nil)
	get := baseimage.GetBaseImagesParams{
		HTTPRequest: r,
	}
	getResponder := api.BaseImageGetBaseImagesHandler.Handle(get, "testCookie")
	var getBody []models.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 3)
}

func TestImageAddImageHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	baseRespBody := addBaseImageEntity(t, api, h, "testBaseImage", "test/base", true, map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get("", "testBaseImage", &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(baseImage.Revision, &baseImage)

	respBody := addImageEntity(t, api, h, "testImage", "testBaseImage", map[string]string{"role": "test"})

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, "testImage", *respBody.Name)
	assert.Equal(t, "testBaseImage", *respBody.BaseImageName)
	assert.Equal(t, *baseRespBody.DockerURL, respBody.DockerURL)
	assert.Equal(t, models.StatusREADY, respBody.Status)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestImageGetImageByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", true, map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get("", "testBaseImage", &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(baseImage.Revision, &baseImage)

	addBody := addImageEntity(t, api, h, "testImage", "testBaseImage", map[string]string{"role": "test"})
	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/image/testImage", nil)
	get := image.GetImageByNameParams{
		HTTPRequest: r,
		ImageName:   "testImage",
	}
	getResponder := api.ImageGetImageByNameHandler.Handle(get, "testCookie")
	var getBody models.Image
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "testImage", *getBody.Name)
	assert.Equal(t, "testBaseImage", *getBody.BaseImageName)
	assert.Equal(t, "test/base", getBody.DockerURL)
	assert.Equal(t, models.StatusREADY, getBody.Status)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)

	r = httptest.NewRequest("GET", "/v1/image/doesNotExist", nil)
	get = image.GetImageByNameParams{
		HTTPRequest: r,
		ImageName:   "doesNotExist",
	}
	getResponder = api.ImageGetImageByNameHandler.Handle(get, "testCookie")
	var errorBody models.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestImageGetImagesHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", true, map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get("", "testBaseImage", &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(baseImage.Revision, &baseImage)

	addImageEntity(t, api, h, "testImage1", "testBaseImage", map[string]string{"role": "test"})
	addImageEntity(t, api, h, "testImage2", "testBaseImage", map[string]string{"role": "test"})
	addImageEntity(t, api, h, "testImage3", "testBaseImage", map[string]string{"role": "test"})

	r := httptest.NewRequest("GET", "/v1/image", nil)
	get := image.GetImagesParams{
		HTTPRequest: r,
	}
	getResponder := api.ImageGetImagesHandler.Handle(get, "testCookie")
	var getBody []models.Image
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 3)
}
