///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

// TODO these tests need to be rewritten for Knative/istio

// import (
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/go-openapi/swag"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
// 	apihandler "github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations/endpoint"
// 	"github.com/vmware/dispatch/pkg/api/v1"
// 	helpers "github.com/vmware/dispatch/pkg/testing/api"
// )

// func assertAPIEqual(t *testing.T, expected *v1.API, real *v1.API) {

// 	assert.Equal(t, expected.Enabled, real.Enabled)
// 	assert.Equal(t, *(expected.Name), *(real.Name))
// 	assert.Equal(t, *(expected.Function), *(real.Function))
// 	assert.Equal(t, expected.Authentication, real.Authentication)
// 	assert.Equal(t, expected.Hosts, real.Hosts)
// 	assert.Equal(t, expected.Uris, real.Uris)
// 	assert.Equal(t, expected.Methods, real.Methods)
// 	assert.Equal(t, expected.Protocols, real.Protocols)
// 	assert.Equal(t, expected.TLS, real.TLS)
// }

// func addAPI(t *testing.T, a *operations.APIManagerAPI, apiModel *v1.API) {

// 	params := apihandler.AddAPIParams{
// 		HTTPRequest:  httptest.NewRequest("POST", "/v1/api", nil),
// 		Body:         apiModel,
// 		XDispatchOrg: testOrgID,
// 	}

// 	responder := a.EndpointAddAPIHandler.Handle(params, "cookie")
// 	var respBody v1.API
// 	helpers.HandlerRequest(t, responder, &respBody, 200)

// 	assertAPIEqual(t, apiModel, &respBody)
// }

// func TestAPIAddAPI(t *testing.T) {

// 	a := operations.NewAPIManagerAPI(nil)
// 	es := helpers.MakeEntityStore(t)
// 	h := NewHandlers(nil, es)

// 	helpers.MakeAPI(t, h.ConfigureHandlers, a)

// 	reqBody := &v1.API{
// 		Name:           swag.String("testAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{"test.com", "vmware.com"},
// 		Uris:           []string{"test", "hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}
// 	addAPI(t, a, reqBody)
// }

// func TestAPIAddAPINoHosts(t *testing.T) {

// 	a := operations.NewAPIManagerAPI(nil)
// 	es := helpers.MakeEntityStore(t)
// 	h := NewHandlers(nil, es)

// 	helpers.MakeAPI(t, h.ConfigureHandlers, a)

// 	reqBody := &v1.API{
// 		Name:           swag.String("testAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{},
// 		Uris:           []string{"test", "/hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}

// 	params := apihandler.AddAPIParams{
// 		HTTPRequest:  httptest.NewRequest("POST", "/v1/api", nil),
// 		Body:         reqBody,
// 		XDispatchOrg: testOrgID,
// 	}

// 	responder := a.EndpointAddAPIHandler.Handle(params, "cookie")
// 	var respBody v1.API
// 	helpers.HandlerRequest(t, responder, &respBody, 200)

// 	expected := []string{"/" + testOrgID + "/test", "/" + testOrgID + "/hello"}
// 	assert.Equal(t, expected, respBody.Uris)
// }

// func TestAPIGetAPIs(t *testing.T) {

// 	a := operations.NewAPIManagerAPI(nil)
// 	es := helpers.MakeEntityStore(t)
// 	h := NewHandlers(nil, es)

// 	helpers.MakeAPI(t, h.ConfigureHandlers, a)

// 	oneAPI := &v1.API{
// 		Name:           swag.String("testAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{"test.com", "vmware.com"},
// 		Uris:           []string{"test", "hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}
// 	anotherAPI := &v1.API{
// 		Name:           swag.String("AnotherAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{"test.com", "vmware.com"},
// 		Uris:           []string{"test", "hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}
// 	addAPI(t, a, oneAPI)
// 	addAPI(t, a, anotherAPI)

// 	params := apihandler.GetApisParams{
// 		HTTPRequest:  httptest.NewRequest("GET", "/v1/api", nil),
// 		XDispatchOrg: testOrgID,
// 	}
// 	responder := a.EndpointGetApisHandler.Handle(params, "cookie")
// 	var respBody []*v1.API
// 	helpers.HandlerRequest(t, responder, &respBody, 200)

// 	assert.Equal(t, 2, len(respBody))
// 	if *respBody[0].Name == *oneAPI.Name {
// 		assertAPIEqual(t, oneAPI, respBody[0])
// 		assertAPIEqual(t, anotherAPI, respBody[1])
// 	} else {
// 		assertAPIEqual(t, oneAPI, respBody[1])
// 		assertAPIEqual(t, anotherAPI, respBody[0])
// 	}
// }

// func TestAPIGetAPI(t *testing.T) {

// 	a := operations.NewAPIManagerAPI(nil)
// 	es := helpers.MakeEntityStore(t)
// 	h := NewHandlers(nil, es)

// 	helpers.MakeAPI(t, h.ConfigureHandlers, a)

// 	oneAPI := &v1.API{
// 		Name:           swag.String("testAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{"test.com", "vmware.com"},
// 		Uris:           []string{"test", "hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}
// 	addAPI(t, a, oneAPI)

// 	params := apihandler.GetAPIParams{
// 		HTTPRequest:  httptest.NewRequest("GET", "/v1/api", nil),
// 		API:          *oneAPI.Name,
// 		XDispatchOrg: testOrgID,
// 	}
// 	responder := a.EndpointGetAPIHandler.Handle(params, "cookie")
// 	var respBody v1.API
// 	helpers.HandlerRequest(t, responder, &respBody, 200)
// 	assertAPIEqual(t, oneAPI, &respBody)
// }

// func TestAPIDeleteAPI(t *testing.T) {

// 	a := operations.NewAPIManagerAPI(nil)
// 	es := helpers.MakeEntityStore(t)
// 	h := NewHandlers(nil, es)

// 	helpers.MakeAPI(t, h.ConfigureHandlers, a)

// 	oneAPI := &v1.API{
// 		Name:           swag.String("testAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{"test.com", "vmware.com"},
// 		Uris:           []string{"test", "hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}
// 	addAPI(t, a, oneAPI)

// 	params := apihandler.DeleteAPIParams{
// 		HTTPRequest:  httptest.NewRequest("GET", "/v1/api", nil),
// 		API:          *oneAPI.Name,
// 		XDispatchOrg: testOrgID,
// 	}
// 	responder := a.EndpointDeleteAPIHandler.Handle(params, "cookie")
// 	var respBody v1.API
// 	helpers.HandlerRequest(t, responder, &respBody, 200)
// 	assertAPIEqual(t, oneAPI, &respBody)
// }

// func TestAPIUpdateAPI(t *testing.T) {

// 	a := operations.NewAPIManagerAPI(nil)
// 	es := helpers.MakeEntityStore(t)
// 	h := NewHandlers(nil, es)

// 	helpers.MakeAPI(t, h.ConfigureHandlers, a)

// 	oneAPI := &v1.API{
// 		Name:           swag.String("testAPI"),
// 		Function:       swag.String("testFunction"),
// 		Authentication: "public",
// 		Enabled:        true,
// 		Hosts:          []string{"test.com", "vmware.com"},
// 		Uris:           []string{"test", "hello"},
// 		Methods:        []string{"GET", "POST"},
// 		Protocols:      []string{"http", "https"},
// 		TLS:            "testtls",
// 	}
// 	addAPI(t, a, oneAPI)

// 	oneAPI.Hosts = []string{"test.com"}
// 	oneAPI.Uris = []string{"anothertest", "anotherhello"}
// 	params := apihandler.UpdateAPIParams{
// 		HTTPRequest:  httptest.NewRequest("GET", "/v1/api", nil),
// 		API:          *oneAPI.Name,
// 		Body:         oneAPI,
// 		XDispatchOrg: testOrgID,
// 	}
// 	responder := a.EndpointUpdateAPIHandler.Handle(params, "cookie")
// 	var respBody v1.API
// 	helpers.HandlerRequest(t, responder, &respBody, 200)
// 	assertAPIEqual(t, oneAPI, &respBody)
// }
