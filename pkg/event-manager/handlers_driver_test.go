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

func addDriverEntity(t *testing.T, api *operations.EventManagerAPI, name, driverType string) *models.Driver {
	reqBody := &models.Driver{
		Name:   swag.String(name),
		Type:   swag.String(driverType),
		Config: []*models.Config{&models.Config{Key: "vcenterurl", Value: "vcenterurl"}},
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivers", nil)
	params := drivers.AddDriverParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverHandler.Handle(params, "testCookie")
	var respBody models.Driver
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func addDriverEntityWithError(t *testing.T, api *operations.EventManagerAPI, name string) *models.Error {
	reqBody := &models.Driver{
		Name: swag.String(name),
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivers", nil)
	params := drivers.AddDriverParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, 400)
	return &respBody
}

func TestDriversAddDriverHandlerError(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverEntityWithError(t, api, "test driver")
	assert.NotEmpty(t, respBody.Message)
	assert.Equal(t, int64(http.StatusBadRequest), respBody.Code)
}

func TestDriversAddDriverHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverEntity(t, api, "drivername", "vcenter")
	assert.Equal(t, "drivername", *respBody.Name)
	assert.Equal(t, "vcenter", *respBody.Type)
}

func TestDriversGetDriverHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addDriverEntity(t, api, "drivername", "vcenter")
	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/event/eventdrivers/drivername", nil)
	get := drivers.GetDriverParams{
		HTTPRequest: r,
		DriverName:  "drivername",
	}
	getResponder := api.DriversGetDriverHandler.Handle(get, "testCookie")
	var getBody models.Driver
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "drivername", *getBody.Name)
	assert.Equal(t, "vcenter", *getBody.Type)

	r = httptest.NewRequest("GET", "/v1/event/eventdrivers/doesNotExist", nil)
	get = drivers.GetDriverParams{
		HTTPRequest: r,
		DriverName:  "doesNotExist",
	}
	getResponder = api.DriversGetDriverHandler.Handle(get, "testCookie")

	var errorBody models.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestDriversDeleteDriverHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{es, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addDriverEntity(t, api, "mydriver", "vcenter")
	assert.NotEmpty(t, addBody.ID)

	r := httptest.NewRequest("GET", "/v1/event/eventdrivers", nil)
	get := drivers.GetDriversParams{
		HTTPRequest: r,
	}
	getResponder := api.DriversGetDriversHandler.Handle(get, "testCookie")
	var getBody []models.Driver
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/events/eventdrivers/mydriver", nil)
	del := drivers.DeleteDriverParams{
		HTTPRequest: r,
		DriverName:  "mydriver",
	}
	delResponder := api.DriversDeleteDriverHandler.Handle(del, "testCookie")
	var delBody models.Driver
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "mydriver", *delBody.Name)

	getResponder = api.DriversGetDriversHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}
