///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/controller"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/functions"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

//go:generate mockery -name ImageGetter -case underscore -dir . -note "CLOSE THIS FILE AS QUICKLY AS POSSIBLE"

const (
	testOrgID = "testOrg"
)

func TestStoreAddFunctionHandler(t *testing.T) {
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	var tags []*v1.Tag
	tags = append(tags, &v1.Tag{Key: "role", Value: "test"})
	schema := v1.Schema{
		In: map[string]interface{}{
			"type":  "object",
			"title": "schema.in",
		},
		Out: map[string]interface{}{
			"type":  "object",
			"title": "schema.out",
		},
	}
	source := &v1.Source{
		Code: []byte("some source"),
	}
	reqBody := &v1.Function{
		Name:   swag.String("testEntity"),
		Schema: &schema,
		Source: source,
		Image:  swag.String("imageID"),
		Tags:   tags,
	}
	r := httptest.NewRequest("POST", "/v1/function", nil)
	params := fnstore.AddFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgID,
	}
	responder := api.StoreAddFunctionHandler.Handle(params, "testCookie")
	var respBody v1.Function
	helpers.HandlerRequest(t, responder, &respBody, 201)

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, reqBody.Name, respBody.Name)
	assert.Equal(t, reqBody.Schema, respBody.Schema)
	assert.Nil(t, respBody.Source)
	assert.Equal(t, reqBody.Image, respBody.Image)
	assert.Len(t, respBody.Tags, 1)
	assert.Equal(t, "role", respBody.Tags[0].Key)
	assert.Equal(t, "test", respBody.Tags[0].Value)
}

func TestHandlers_addFunction_duplicate(t *testing.T) {
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	source := &v1.Source{
		Code: []byte("some source"),
	}
	reqBody := &v1.Function{
		Name:   swag.String("testEntity"),
		Source: source,
		Image:  swag.String("imageID"),
	}
	r := httptest.NewRequest("POST", "/v1/function", nil)
	params := fnstore.AddFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgID,
	}
	responder := api.StoreAddFunctionHandler.Handle(params, "testCookie")
	var respBody v1.Function
	helpers.HandlerRequest(t, responder, &respBody, 201)

	f := new(functions.Function)
	err := handlers.Store.Get(context.Background(), testOrgID, *reqBody.Name, entitystore.Options{}, f)
	require.NoError(t, err)
	assert.Equal(t, *reqBody.Name, f.Name)

	responder = api.StoreAddFunctionHandler.Handle(params, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 409)

	var sources []*functions.Source
	err = handlers.Store.List(context.Background(), testOrgID, entitystore.Options{}, &sources)
	require.NoError(t, err)
	assert.Equal(t, 2, len(sources))
	for _, source := range sources {
		assert.Equal(t, *reqBody.Name, source.Function)
		if source.Name != f.SourceName {
			assert.Equal(t, entitystore.StatusDELETING, source.Status)
		}
	}
}

func TestHandlers_runFunction_notREADY(t *testing.T) {
	store := helpers.MakeEntityStore(t)
	watcher := make(chan controller.WatchEvent, 1)
	handlers := &Handlers{
		Watcher: watcher,
		Store:   store,
	}

	testFuncName := "testFunction"

	handlers.Store.Add(context.Background(), &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:           testFuncName,
			Status:         entitystore.StatusCREATING,
			OrganizationID: testOrgID,
		},
		// other fields are unimportant for this test
	})

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("POST", fmt.Sprintf("/v1/runs?functionName=%s", testFuncName), nil)
	reqBody := &v1.Run{}
	params := fnrunner.RunFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		FunctionName: &testFuncName,
		XDispatchOrg: testOrgID,
	}
	responder := api.RunnerRunFunctionHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, 404)

	assert.EqualValues(t, http.StatusNotFound, respBody.Code)
	assert.Equal(t, "function testFunction is not READY", *respBody.Message)
	assert.Len(t, watcher, 0)
}

func TestHandlers_runFunction_READY(t *testing.T) {
	store := helpers.MakeEntityStore(t)
	watcher := make(chan controller.WatchEvent, 1)
	handlers := &Handlers{
		Watcher: watcher,
		Store:   store,
	}

	testFuncName := "testFunction"

	function := &functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name:           testFuncName,
			Status:         entitystore.StatusREADY,
			OrganizationID: testOrgID,
		},
		// other fields are unimportant for this test
	}
	store.Add(context.Background(), function)

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("POST", fmt.Sprintf("/v1/runs?functionName=%s", testFuncName), nil)
	reqBody := &v1.Run{}
	params := fnrunner.RunFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		FunctionName: &testFuncName,
		XDispatchOrg: testOrgID,
	}
	responder := api.RunnerRunFunctionHandler.Handle(params, "testCookie")
	var respBody v1.Run
	helpers.HandlerRequest(t, responder, &respBody, 202)

	assert.Equal(t, testFuncName, respBody.FunctionName)
	assert.EqualValues(t, entitystore.StatusINITIALIZED, respBody.Status)
	assert.Equal(t, runEntityToModel((<-watcher).Entity.(*functions.FnRun)), &respBody)
}

func TestHandlers_getRuns(t *testing.T) {
	store := helpers.MakeEntityStore(t)
	handlers := &Handlers{
		Store: store,
	}

	testFuncName := "testFunction"
	diffFuncName := "diffFunction"

	run1 := &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name:           "run1",
			OrganizationID: testOrgID,
		},
		FunctionName: testFuncName,
	}
	run2 := &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name:           "run2",
			OrganizationID: testOrgID,
		},
		FunctionName: diffFuncName,
	}
	run3 := &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			Name:           "run3",
			OrganizationID: testOrgID,
		},
		FunctionName: testFuncName,
	}

	store.Add(context.Background(), run1)
	store.Add(context.Background(), run2)

	time.Sleep(time.Second)
	now := time.Now()
	store.Add(context.Background(), run3)

	api := operations.NewFunctionManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("GET", "/v1/runs", nil)
	params := fnrunner.GetRunsParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgID,
	}
	responder := api.RunnerGetRunsHandler.Handle(params, "testcookie")
	var respBody []v1.Run
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.Equal(t, 3, len(respBody))

	r = httptest.NewRequest("GET", fmt.Sprintf("/v1/runs?functionName=%s", testFuncName), nil)
	params = fnrunner.GetRunsParams{
		HTTPRequest:  r,
		FunctionName: &testFuncName,
		XDispatchOrg: testOrgID,
	}
	responder = api.RunnerGetRunsHandler.Handle(params, "testcookie")
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.Equal(t, 2, len(respBody))

	afterSecs := now.Unix()
	r = httptest.NewRequest("GET", fmt.Sprintf("/v1/runs?functionName=%s?since=%d", testFuncName, afterSecs), nil)
	params = fnrunner.GetRunsParams{
		HTTPRequest:  r,
		FunctionName: &testFuncName,
		Since:        &afterSecs,
		XDispatchOrg: testOrgID,
	}
	responder = api.RunnerGetRunsHandler.Handle(params, "testcookie")
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.Equal(t, 1, len(respBody))
	assert.EqualValues(t, run3.Name, respBody[0].Name)
}

func TestStoreGetFunctionHandler(t *testing.T) {
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewFunctionManagerAPI(nil)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	var tags []*v1.Tag
	tags = append(tags, &v1.Tag{Key: "role", Value: "test"})
	schema := v1.Schema{
		In: map[string]interface{}{
			"type":  "object",
			"title": "schema.in",
		},
		Out: map[string]interface{}{
			"type":  "object",
			"title": "schema.out",
		},
	}
	source := &v1.Source{
		Code: []byte("some source"),
	}
	reqBody := &v1.Function{
		Name:   swag.String("testEntity"),
		Schema: &schema,
		Source: source,
		Image:  swag.String("imageID"),
		Tags:   tags,
	}
	r := httptest.NewRequest("POST", "/v1/function", nil)
	add := fnstore.AddFunctionParams{
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgID,
	}
	addResponder := api.StoreAddFunctionHandler.Handle(add, "testCookie")
	var addBody v1.Function
	helpers.HandlerRequest(t, addResponder, &addBody, 201)

	assert.NotEmpty(t, addBody.ID)

	id := addBody.ID
	createdTime := addBody.CreatedTime
	r = httptest.NewRequest("GET", "/v1/function/testEntity", nil)
	get := fnstore.GetFunctionParams{
		HTTPRequest:  r,
		FunctionName: "testEntity",
		XDispatchOrg: testOrgID,
	}
	getResponder := api.StoreGetFunctionHandler.Handle(get, "testCookie")
	var getBody v1.Function
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, id, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, reqBody.Name, getBody.Name)
	assert.Equal(t, reqBody.Schema, getBody.Schema)
	assert.Nil(t, getBody.Source)
	assert.Equal(t, reqBody.Schema, getBody.Schema)
	assert.Len(t, getBody.Tags, 1)
	assert.Equal(t, "role", getBody.Tags[0].Key)
	assert.Equal(t, "test", getBody.Tags[0].Value)
}

func Test_runModelToEntitySecret(t *testing.T) {
	runModel0 := v1.Run{Secrets: []string{}}
	bs, _ := json.Marshal(runModel0)
	secrets := []string{"x", "y", "z"}
	f := functions.Function{Secrets: secrets}
	var runModel v1.Run
	json.Unmarshal(bs, &runModel)
	assert.NotNil(t, runModel.Secrets)
	fnRun := runModelToEntity(&runModel, &f)
	assert.Equal(t, secrets, fnRun.Secrets)
}
