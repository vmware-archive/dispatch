///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func testHandlers(store entitystore.EntityStore) *Handlers {
	return &Handlers{
		store: store,
	}
}

func addDriverEntity(t *testing.T, api *operations.EventManagerAPI, name, driverType string) *v1.EventDriver {
	reqBody := &v1.EventDriver{
		Name:   swag.String(name),
		Type:   swag.String(driverType),
		Config: []*v1.Config{&v1.Config{Key: "vcenterurl", Value: "vcenterurl"}},
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivers", nil)
	params := drivers.AddDriverParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverHandler.Handle(params, "testCookie")
	var respBody v1.EventDriver
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func addDriverEntityWithError(t *testing.T, api *operations.EventManagerAPI, name string) *v1.Error {
	reqBody := &v1.EventDriver{
		Name: swag.String(name),
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivers", nil)
	params := drivers.AddDriverParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, 400)
	return &respBody
}

func addDriverTypeEntity(t *testing.T, api *operations.EventManagerAPI, name, image string) *v1.EventDriverType {
	reqBody := &v1.EventDriverType{
		Name:  swag.String(name),
		Image: swag.String(image),
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivertypes", nil)
	params := drivers.AddDriverTypeParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverTypeHandler.Handle(params, "testCookie")
	var respBody v1.EventDriverType
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func addDriverTypeEntityWithError(t *testing.T, api *operations.EventManagerAPI, name string) *v1.Error {
	reqBody := &v1.EventDriverType{
		Name: swag.String(name),
	}
	r := httptest.NewRequest("POST", "/v1/event/eventdrivertypes", nil)
	params := drivers.AddDriverTypeParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.DriversAddDriverTypeHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, 400)
	return &respBody
}

func TestDriversAddDriverHandlerError(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverEntityWithError(t, api, "test driver")
	assert.NotEmpty(t, respBody.Message)
	assert.Equal(t, int64(http.StatusBadRequest), respBody.Code)
}

func TestDriversAddDriverHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverEntity(t, api, "drivername", "vcenter")
	assert.Equal(t, "drivername", *respBody.Name)
	assert.Equal(t, "vcenter", *respBody.Type)
}

func TestDriversGetDriverHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
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
	var getBody v1.EventDriver
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

	var errorBody v1.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestDriversDeleteDriverHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addDriverEntity(t, api, "mydriver", "vcenter")
	assert.NotEmpty(t, addBody.ID)

	r := httptest.NewRequest("GET", "/v1/event/eventdrivers", nil)
	get := drivers.GetDriversParams{
		HTTPRequest: r,
	}
	getResponder := api.DriversGetDriversHandler.Handle(get, "testCookie")
	var getBody []v1.EventDriver
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/events/eventdrivers/mydriver", nil)
	del := drivers.DeleteDriverParams{
		HTTPRequest: r,
		DriverName:  "mydriver",
	}
	delResponder := api.DriversDeleteDriverHandler.Handle(del, "testCookie")
	var delBody v1.EventDriver
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "mydriver", *delBody.Name)

	getResponder = api.DriversGetDriversHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}

func TestDriversAddDriverTypeHandlerError(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverTypeEntityWithError(t, api, "test driver")
	assert.NotEmpty(t, respBody.Message)
	assert.Equal(t, int64(http.StatusBadRequest), respBody.Code)
}

func TestDriversAddDriverTypeHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addDriverTypeEntity(t, api, "typename", "golang:latest")
	assert.Equal(t, "typename", *respBody.Name)
	assert.Equal(t, "golang:latest", *respBody.Image)
}

func TestDriversGetDriverTypeHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
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
	var getBody v1.EventDriverType
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

	var errorBody v1.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestDriversDeleteDriverTypeHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := testHandlers(es)
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	r := httptest.NewRequest("GET", "/v1/event/eventdrivertypes", nil)
	get := drivers.GetDriverTypesParams{
		HTTPRequest: r,
	}
	getResponder := api.DriversGetDriverTypesHandler.Handle(get, "testCookie")
	var getBody []v1.EventDriver
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 1)

	addBody := addDriverTypeEntity(t, api, "typename", "golang:latest")
	assert.NotEmpty(t, addBody.ID)

	getResponder = api.DriversGetDriverTypesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 2)

	r = httptest.NewRequest("DELETE", "/v1/events/eventdrivertypes/typename", nil)
	del := drivers.DeleteDriverTypeParams{
		HTTPRequest:    r,
		DriverTypeName: "typename",
	}
	delResponder := api.DriversDeleteDriverTypeHandler.Handle(del, "testCookie")
	var delBody v1.EventDriverType
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "typename", *delBody.Name)

	getResponder = api.DriversGetDriverTypesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}
