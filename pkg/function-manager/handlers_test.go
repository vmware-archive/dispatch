///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/functions"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

//go:generate mockery -name ImageManager -case underscore -dir .

func TestStoreAddFunctionHandler(t *testing.T) {
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

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

func TestHandlers_runFunction_notREADY(t *testing.T) {
	store := helpers.MakeEntityStore(t)
	watcher := make(chan entitystore.Entity, 1)
	handlers := &Handlers{
		Watcher: watcher,
		Store:   store,
	}

	testFuncName := "testFunction"

	handlers.Store.Add(&functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:   testFuncName,
			Status: entitystore.StatusCREATING,
		},
		// other fields are unimportant for this test
	})

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("POST", fmt.Sprintf("/v1/runs?functionName=%s", testFuncName), nil)
	reqBody := &models.Run{}
	params := fnrunner.RunFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		FunctionName: &testFuncName,
	}
	responder := api.RunnerRunFunctionHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, 404)

	assert.EqualValues(t, http.StatusNotFound, respBody.Code)
	assert.Equal(t, "function is not READY", *respBody.Message)
	assert.Len(t, watcher, 0)
}

func TestHandlers_runFunction_READY(t *testing.T) {
	store := helpers.MakeEntityStore(t)
	watcher := make(chan entitystore.Entity, 1)
	handlers := &Handlers{
		Watcher: watcher,
		Store:   store,
	}

	testFuncName := "testFunction"

	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:   testFuncName,
			Status: entitystore.StatusREADY,
		},
		// other fields are unimportant for this test
	}
	store.Add(function)

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("POST", fmt.Sprintf("/v1/runs?functionName=%s", testFuncName), nil)
	reqBody := &models.Run{}
	params := fnrunner.RunFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		FunctionName: &testFuncName,
	}
	responder := api.RunnerRunFunctionHandler.Handle(params, "testCookie")
	var respBody models.Run
	helpers.HandlerRequest(t, responder, &respBody, 202)

	assert.Equal(t, testFuncName, respBody.FunctionName)
	assert.EqualValues(t, entitystore.StatusINITIALIZED, respBody.Status)
	assert.Equal(t, runEntityToModel((<-watcher).(*functions.FnRun)), &respBody)
}

func TestStoreGetFunctionHandler(t *testing.T) {
	handlers := &Handlers{
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
	f := functions.Function{Secrets: secrets}
	var runModel models.Run
	json.Unmarshal(bs, &runModel)
	assert.NotNil(t, runModel.Secrets)
	fnRun := runModelToEntity(&runModel, &f)
	assert.Equal(t, secrets, fnRun.Secrets)
}
