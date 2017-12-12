///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/function-manager/mocks"
	"github.com/vmware/dispatch/pkg/functions"
	fnmocks "github.com/vmware/dispatch/pkg/functions/mocks"
	"github.com/vmware/dispatch/pkg/functions/runner"
	"github.com/vmware/dispatch/pkg/functions/validator"
	image "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	imagemodels "github.com/vmware/dispatch/pkg/image-manager/gen/models"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

//go:generate mockery -name ImageManager -case underscore -dir .

type testEntity struct {
	entitystore.BaseEntity
	Value string `json:"value"`
}

func TestStoreAddFunctionHandler(t *testing.T) {
	imgMgr := &mocks.ImageManager{}
	imgMgr.On("GetImageByName", mock.Anything, mock.Anything).Return(
		&image.GetImageByNameOK{
			Payload: &imagemodels.Image{
				DockerURL: "test/image:latest",
				Language:  imagemodels.LanguagePython3,
			},
		}, nil)
	faas := &fnmocks.FaaSDriver{}
	faas.On("Create", "testEntity", &functions.Exec{
		Code: "some code", Main: "main", Image: "test/image:latest", Language: "python3",
	}).Return(nil)
	handlers := &Handlers{
		FaaS: faas,
		Runner: runner.New(&runner.Config{
			Faas:      faas,
			Validator: validator.NoOp(),
		}),
		Store:     helpers.MakeEntityStore(t),
		ImgClient: imgMgr,
	}

	api := operations.NewFunctionManagerAPI(nil)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	var tags []*models.Tag
	tags = append(tags, &models.Tag{Key: "role", Value: "test"})
	schema := models.Schema{
		In: map[string]interface{}{
			"type":  "object",
			"title": "schema.in",
		},
		Out: map[string]interface{}{
			"type":  "object",
			"title": "schema.out",
		},
	}
	reqBody := &models.Function{
		Name:   swag.String("testEntity"),
		Schema: &schema,
		Code:   swag.String("some code"),
		Image:  swag.String("imageID"),
		Tags:   tags,
	}
	r := httptest.NewRequest("POST", "/v1/function", nil)
	params := fnstore.AddFunctionParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.StoreAddFunctionHandler.Handle(params, "testCookie")
	var respBody models.Function
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, reqBody.Name, respBody.Name)
	assert.Equal(t, reqBody.Schema, respBody.Schema)
	assert.Equal(t, reqBody.Code, respBody.Code)
	assert.Equal(t, reqBody.Image, respBody.Image)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestStoreGetFunctionHandler(t *testing.T) {
	imgMgr := &mocks.ImageManager{}
	imgMgr.On("GetImageByName", mock.Anything, mock.Anything).Return(
		&image.GetImageByNameOK{
			Payload: &imagemodels.Image{
				DockerURL: "test/image:latest",
				Language:  imagemodels.LanguagePython3,
			},
		}, nil)
	faas := &fnmocks.FaaSDriver{}
	faas.On("Create", "testEntity", &functions.Exec{
		Code: "some code", Main: "main", Image: "test/image:latest", Language: "python3",
	}).Return(nil)
	handlers := &Handlers{
		FaaS: faas,
		Runner: runner.New(&runner.Config{
			Faas:      faas,
			Validator: validator.NoOp(),
		}),
		Store:     helpers.MakeEntityStore(t),
		ImgClient: imgMgr,
	}

	api := operations.NewFunctionManagerAPI(nil)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	var tags []*models.Tag
	tags = append(tags, &models.Tag{Key: "role", Value: "test"})
	schema := models.Schema{
		In: map[string]interface{}{
			"type":  "object",
			"title": "schema.in",
		},
		Out: map[string]interface{}{
			"type":  "object",
			"title": "schema.out",
		},
	}
	reqBody := &models.Function{
		Name:   swag.String("testEntity"),
		Schema: &schema,
		Code:   swag.String("some code"),
		Image:  swag.String("imageID"),
		Tags:   tags,
	}
	r := httptest.NewRequest("POST", "/v1/function", nil)
	add := fnstore.AddFunctionParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	addResponder := api.StoreAddFunctionHandler.Handle(add, "testCookie")
	var addBody models.Function
	helpers.HandlerRequest(t, addResponder, &addBody, 200)

	assert.NotEmpty(t, addBody.ID)

	id := addBody.ID
	createdTime := addBody.CreatedTime
	r = httptest.NewRequest("GET", "/v1/function/testEntity", nil)
	get := fnstore.GetFunctionParams{
		HTTPRequest:  r,
		FunctionName: "testEntity",
	}
	getResponder := api.StoreGetFunctionHandler.Handle(get, "testCookie")
	var getBody models.Function
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, id, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, reqBody.Name, getBody.Name)
	assert.Equal(t, reqBody.Schema, getBody.Schema)
	assert.Equal(t, reqBody.Code, getBody.Code)
	assert.Equal(t, reqBody.Schema, getBody.Schema)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)
}

func Test_runModelToEntitySecret(t *testing.T) {
	runModel0 := models.Run{Secrets: []string{}}
	bs, _ := json.Marshal(runModel0)
	secrets := []string{"x", "y", "z"}
	f := Function{Secrets: secrets}
	var runModel models.Run
	json.Unmarshal(bs, &runModel)
	assert.NotNil(t, runModel.Secrets)
	fnRun := runModelToEntity(&runModel, &f)
	assert.Equal(t, secrets, fnRun.Secrets)
}
