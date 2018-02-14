///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func addDriverTypeEntity(t *testing.T, api *operations.EventManagerAPI, name, image string) *models.DriverType {
	reqBody := &models.DriverType{
		Name:  swag.String(name),
		Image: swag.String(image),
		Mode:  swag.String("http"),
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivertypes", nil)
	params := drivers.AddDriverTypeParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverTypeHandler.Handle(params, "testCookie")
	var respBody models.DriverType
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func addDriverTypeEntityWithError(t *testing.T, api *operations.EventManagerAPI, name string) *models.Error {
	reqBody := &models.DriverType{
		Name: swag.String(name),
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivertypes", nil)
	params := drivers.AddDriverTypeParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverTypeHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, 400)
	return &respBody
}

func TestDriversAddDriverTypeHandlerError(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverTypeEntityWithError(t, api, "test driver")
	assert.NotEmpty(t, respBody.Message)
	assert.Equal(t, int64(http.StatusBadRequest), respBody.Code)
}

func TestDriversAddDriverTypeHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverTypeEntity(t, api, "typename", "golang:latest")
	assert.Equal(t, "typename", *respBody.Name)
	assert.Equal(t, "golang:latest", *respBody.Image)
}

func TestDriversGetDriverTypeHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addDriverTypeEntity(t, api, "typename", "golang:latest")
	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/event/eventdrivertypes/drivertypename", nil)
	get := drivers.GetDriverTypeParams{
		HTTPRequest:    r,
		DriverTypeName: "typename",
	}
	getResponder := api.DriversGetDriverTypeHandler.Handle(get, "testCookie")
	var getBody models.DriverType
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "typename", *getBody.Name)
	assert.Equal(t, "golang:latest", *getBody.Image)

	r = httptest.NewRequest("GET", "/v1/event/eventdrivertypes/doesNotExist", nil)
	get = drivers.GetDriverTypeParams{
		HTTPRequest:    r,
		DriverTypeName: "doesNotExist",
	}
	getResponder = api.DriversGetDriverTypeHandler.Handle(get, "testCookie")

	var errorBody models.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestDriversDeleteDriverTypeHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addDriverTypeEntity(t, api, "typename", "golang:latest")
	assert.NotEmpty(t, addBody.ID)

	r := httptest.NewRequest("GET", "/v1/event/eventdrivertypes", nil)
	get := drivers.GetDriverTypesParams{
		HTTPRequest: r,
	}
	getResponder := api.DriversGetDriverTypesHandler.Handle(get, "testCookie")
	var getBody []models.Driver
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/events/eventdrivertypes/typename", nil)
	del := drivers.DeleteDriverTypeParams{
		HTTPRequest:    r,
		DriverTypeName: "typename",
	}
	delResponder := api.DriversDeleteDriverTypeHandler.Handle(del, "testCookie")
	var delBody models.DriverType
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "typename", *delBody.Name)

	getResponder = api.DriversGetDriverTypesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}
