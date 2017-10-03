///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package functionmanager

import (
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations"
	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/mocks"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/runner"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/validator"
	helpers "gitlab.eng.vmware.com/serverless/serverless/pkg/testing/api"
)

type testEntity struct {
	entitystore.BaseEntity
	Value string `json:"value"`
}

func TestStoreAddFunctionHandler(t *testing.T) {
	faas := &mocks.FaaSDriver{}
	faas.On("Create", "testEntity", &functions.Exec{
		Code: "some code", Main: "main", Image: "imageID",
	}).Return(nil)
	handlers := &Handlers{
		FaaS: faas,
		Runner: runner.New(&runner.Config{
			Faas:      faas,
			Validator: validator.NoOp(),
		}),
		Store: helpers.MakeEntityStore(t),
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
	responder := api.StoreAddFunctionHandler.Handle(params)
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

func TestStoreGetFunctionByNameHandler(t *testing.T) {
	faas := &mocks.FaaSDriver{}
	faas.On("Create", "testEntity", &functions.Exec{
		Code: "some code", Main: "main", Image: "imageID",
	}).Return(nil)
	handlers := &Handlers{
		FaaS: faas,
		Runner: runner.New(&runner.Config{
			Faas:      faas,
			Validator: validator.NoOp(),
		}),
		Store: helpers.MakeEntityStore(t),
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
	addResponder := api.StoreAddFunctionHandler.Handle(add)
	var addBody models.Function
	helpers.HandlerRequest(t, addResponder, &addBody, 200)

	assert.NotEmpty(t, addBody.ID)

	id := addBody.ID
	createdTime := addBody.CreatedTime
	r = httptest.NewRequest("GET", "/v1/function/testEntity", nil)
	get := fnstore.GetFunctionByNameParams{
		HTTPRequest:  r,
		FunctionName: "testEntity",
	}
	getResponder := api.StoreGetFunctionByNameHandler.Handle(get)
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
