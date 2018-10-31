///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/api/v1"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
	baseimage "github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations/image"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

const (
	testOrgID = "testOrg"
)

func addBaseImageEntity(t *testing.T, api *operations.ImageManagerAPI, h *Handlers, name, dockerURL, language string, tags map[string]string) *v1.BaseImage {
	var entityTags []*v1.Tag
	for k, v := range tags {
		entityTags = append(entityTags, &v1.Tag{Key: k, Value: v})
	}

	reqBody := &v1.BaseImage{
		Name:      swag.String(name),
		DockerURL: swag.String(dockerURL),
		Language:  swag.String(language),
		Tags:      entityTags,
	}
	r := httptest.NewRequest("POST", "/v1/baseimage", nil)
	params := baseimage.AddBaseImageParams{
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgID,
	}
	responder := api.BaseImageAddBaseImageHandler.Handle(params, "testCookie")
	var respBody v1.BaseImage
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func addImageEntity(t *testing.T, api *operations.ImageManagerAPI, h *Handlers, name, baseImageName string, tags map[string]string) *v1.Image {
	var entityTags []*v1.Tag
	for k, v := range tags {
		entityTags = append(entityTags, &v1.Tag{Key: k, Value: v})
	}

	reqBody := &v1.Image{
		Name:          swag.String(name),
		BaseImageName: swag.String(baseImageName),
		Tags:          entityTags,
	}
	r := httptest.NewRequest("POST", "/v1/image", nil)
	params := image.AddImageParams{
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgID,
	}
	responder := api.ImageAddImageHandler.Handle(params, "testCookie")
	var respBody v1.Image
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func TestBaseImageAddBaseImageHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addBaseImageEntity(t, api, h, "testEntity", "test/base", "python3", map[string]string{"role": "test"})

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, "testEntity", *respBody.Name)
	assert.Equal(t, "test/base", *respBody.DockerURL)
	assert.Equal(t, v1.StatusINITIALIZED, respBody.Status)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestBaseImageGetBaseImageByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addBaseImageEntity(t, api, h, "testEntity", "test/base", "python3", map[string]string{"role": "test"})

	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/baseimage/testEntity", nil)
	get := baseimage.GetBaseImageByNameParams{
		HTTPRequest:   r,
		BaseImageName: "testEntity",
		XDispatchOrg:  testOrgID,
	}
	getResponder := api.BaseImageGetBaseImageByNameHandler.Handle(get, "testCookie")
	var getBody v1.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "testEntity", *getBody.Name)
	assert.Equal(t, "test/base", *getBody.DockerURL)
	assert.Equal(t, v1.StatusINITIALIZED, getBody.Status)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)

	r = httptest.NewRequest("GET", "/v1/baseimage/doesNotExist", nil)
	get = baseimage.GetBaseImageByNameParams{
		HTTPRequest:   r,
		BaseImageName: "doesNotExist",
	}
	getResponder = api.BaseImageGetBaseImageByNameHandler.Handle(get, "testCookie")

	var errorBody v1.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestBaseImageGetBaseImagesHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testEntity1", "test/base", "python3", map[string]string{"role": "test", "item": "1"})
	addBaseImageEntity(t, api, h, "testEntity2", "test/base", "python3", map[string]string{"role": "test", "item": "2"})
	addBaseImageEntity(t, api, h, "testEntity3", "test/base", "python3", map[string]string{"role": "test", "item": "3"})

	r := httptest.NewRequest("GET", "/v1/baseimage", nil)
	get := baseimage.GetBaseImagesParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgID,
	}
	getResponder := api.BaseImageGetBaseImagesHandler.Handle(get, "testCookie")
	var getBody []v1.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 3)
}

func TestBaseImageDeleteBaseImageByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testEntity", "test/base", "python3", map[string]string{"role": "test"})

	r := httptest.NewRequest("GET", "/v1/baseimage", nil)
	get := baseimage.GetBaseImagesParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgID,
	}
	getResponder := api.BaseImageGetBaseImagesHandler.Handle(get, "testCookie")
	var getBody []v1.BaseImage
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/baseimage/testEntity", nil)
	del := baseimage.DeleteBaseImageByNameParams{
		HTTPRequest:   r,
		BaseImageName: "testEntity",
		XDispatchOrg:  testOrgID,
	}
	delResponder := api.BaseImageDeleteBaseImageByNameHandler.Handle(del, "testCookie")
	var delBody v1.BaseImage
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "testEntity", *delBody.Name)
	// Status should be unchanged as the actual deletion is done asynchronously
	assert.Equal(t, v1.StatusINITIALIZED, delBody.Status)

	getResponder = api.BaseImageGetBaseImagesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}

func TestImageAddImageHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", "python3", map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get(context.Background(), testOrgID, "testBaseImage", entitystore.Options{}, &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(context.Background(), baseImage.Revision, &baseImage)

	respBody := addImageEntity(t, api, h, "testImage", "testBaseImage", map[string]string{"role": "test"})

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, "testImage", *respBody.Name)
	assert.Equal(t, "testBaseImage", *respBody.BaseImageName)
	assert.Equal(t, "", respBody.DockerURL)
	assert.Equal(t, v1.StatusINITIALIZED, respBody.Status)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestImageGetImageByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", "python3", map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get(context.Background(), testOrgID, "testBaseImage", entitystore.Options{}, &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(context.Background(), baseImage.Revision, &baseImage)

	addBody := addImageEntity(t, api, h, "testImage", "testBaseImage", map[string]string{"role": "test"})
	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/image/testImage", nil)
	get := image.GetImageByNameParams{
		HTTPRequest:  r,
		ImageName:    "testImage",
		XDispatchOrg: testOrgID,
	}
	getResponder := api.ImageGetImageByNameHandler.Handle(get, "testCookie")
	var getBody v1.Image
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "testImage", *getBody.Name)
	assert.Equal(t, "testBaseImage", *getBody.BaseImageName)
	assert.Equal(t, "", getBody.DockerURL)
	assert.Equal(t, v1.StatusINITIALIZED, getBody.Status)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)

	r = httptest.NewRequest("GET", "/v1/image/doesNotExist", nil)
	get = image.GetImageByNameParams{
		HTTPRequest:  r,
		ImageName:    "doesNotExist",
		XDispatchOrg: testOrgID,
	}
	getResponder = api.ImageGetImageByNameHandler.Handle(get, "testCookie")
	var errorBody v1.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestImageGetImagesHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", "python3", map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get(context.Background(), testOrgID, "testBaseImage", entitystore.Options{}, &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(context.Background(), baseImage.Revision, &baseImage)

	addImageEntity(t, api, h, "testImage1", "testBaseImage", map[string]string{"role": "test"})
	addImageEntity(t, api, h, "testImage2", "testBaseImage", map[string]string{"role": "test"})
	addImageEntity(t, api, h, "testImage3", "testBaseImage", map[string]string{"role": "test"})

	r := httptest.NewRequest("GET", "/v1/image", nil)
	get := image.GetImagesParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgID,
	}
	getResponder := api.ImageGetImagesHandler.Handle(get, "testCookie")
	var getBody []v1.Image
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 3)
}

func TestImageUpdateImageByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", "python3", map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get(context.Background(), testOrgID, "testBaseImage", entitystore.Options{}, &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(context.Background(), baseImage.Revision, &baseImage)

	addImageEntity(t, api, h, "testImage", "testBaseImage", map[string]string{"role": "test"})

	r := httptest.NewRequest("GET", "/v1/image", nil)
	get := image.GetImagesParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgID,
	}
	getResponder := api.ImageGetImagesHandler.Handle(get, "testCookie")
	var getBody []v1.Image
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("PUT", "/v1/image/testImage", nil)
	imageName := "testImage"
	baseImageName := "testBaseImage"
	update := image.UpdateImageByNameParams{
		HTTPRequest: r,
		ImageName:   "testImage",
		Body: &v1.Image{
			Name:          &imageName,
			BaseImageName: &baseImageName,
		},
		XDispatchOrg: testOrgID,
	}
	updateReponder := api.ImageUpdateImageByNameHandler.Handle(update, "testCookie")
	var updateBody v1.Image
	helpers.HandlerRequest(t, updateReponder, &updateBody, 200)
	assert.Equal(t, "testImage", *updateBody.Name)
	assert.Equal(t, 0, len(updateBody.Tags))
}

func TestImageDeleteImagesByNameHandler(t *testing.T) {
	api := operations.NewImageManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := NewHandlers(nil, nil, nil, es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBaseImageEntity(t, api, h, "testBaseImage", "test/base", "python3", map[string]string{"role": "test"})

	baseImage := BaseImage{}
	err := es.Get(context.Background(), testOrgID, "testBaseImage", entitystore.Options{}, &baseImage)
	assert.NoError(t, err)
	baseImage.Status = StatusREADY
	_, err = es.Update(context.Background(), baseImage.Revision, &baseImage)

	addImageEntity(t, api, h, "testImage", "testBaseImage", map[string]string{"role": "test"})

	r := httptest.NewRequest("GET", "/v1/image", nil)
	get := image.GetImagesParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgID,
	}
	getResponder := api.ImageGetImagesHandler.Handle(get, "testCookie")
	var getBody []v1.Image
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/image/testImage", nil)
	del := image.DeleteImageByNameParams{
		HTTPRequest:  r,
		ImageName:    "testImage",
		XDispatchOrg: testOrgID,
	}
	delResponder := api.ImageDeleteImageByNameHandler.Handle(del, "testCookie")
	var delBody v1.Image
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "testImage", *delBody.Name)
	// Status should be unchanged as the actual deletion is done asynchronously
	assert.Equal(t, v1.StatusDELETED, delBody.Status)

	getResponder = api.ImageGetImagesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}
