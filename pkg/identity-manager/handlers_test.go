///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"

	"github.com/vmware/dispatch/pkg/entity-store"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func addTestData(store entitystore.EntityStore) {
	// Add test policies and rules
	e := Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           "test-policy-1",
		},
		Rules: []Rule{
			Rule{
				Subjects:  []string{"readonly-user@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"get"},
			},
			Rule{
				Subjects:  []string{"super-admin@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			}},
	}
	store.Add(&e)
}

func TestHomeHandler(t *testing.T) {
	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	params := operations.HomeParams{
		HTTPRequest: httptest.NewRequest("GET", "/v1/iam/home", nil),
	}
	responder := api.HomeHandler.Handle(params, nil)

	var respBody models.Message
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)
	assert.Equal(t, "Home Page, You have already logged in", *respBody.Message)
}

func TestRootHandler(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	params := operations.RootParams{
		HTTPRequest: httptest.NewRequest("GET", "/", nil),
	}
	responder := api.RootHandler.Handle(params)

	var respBody models.Message
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)
	assert.Equal(t, "Default Root Page", *respBody.Message)
}

func TestAuthHandlerPolicyPass(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	addTestData(es)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "GET")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerWithoutPolicyData(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "GET")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerNonResourcePass(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	addTestData(es)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/echo")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "GET")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "noname@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerPolicyFail(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	addTestData(es)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "POST")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyNoValidHeader(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	request := httptest.NewRequest("GET", "/auth", nil)
	// Missing Email Header
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "POST")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestGetRequestAttributesNoEmailHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "POST")
	_, err := getRequestAttributes(request)
	assert.EqualError(t, err, "X-Forwarded-Email header not found")
}

func TestGetRequestAttributesNoMethodHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "super-admin@example.com")
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	_, err := getRequestAttributes(request)
	assert.EqualError(t, err, "X-Original-Method header not found")
}

func TestGetRequestAttributesNoURLHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "super-admin@example.com")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "POST")
	_, err := getRequestAttributes(request)
	assert.EqualError(t, err, "X-Auth-Request-Redirect header not found")
}

func TestGetRequestAttributesValidResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/v1/function")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "POST")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "super-admin@example.com")
	attrRecord, _ := getRequestAttributes(request)
	assert.Equal(t, "super-admin@example.com", attrRecord.userEmail)
	assert.Equal(t, ActionCreate, attrRecord.action)
	assert.Equal(t, "function", attrRecord.resource)
	assert.Equal(t, true, attrRecord.isResourceRequest)
	assert.Equal(t, "", attrRecord.path)
}

func TestGetRequestAttributesNonResourcePath(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "/echo")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "GET")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "super-admin@example.com")
	attrRecord, _ := getRequestAttributes(request)
	assert.Equal(t, "super-admin@example.com", attrRecord.userEmail)
	assert.Equal(t, ActionGet, attrRecord.action)
	assert.Equal(t, "", attrRecord.resource)
	assert.Equal(t, false, attrRecord.isResourceRequest)
	assert.Equal(t, "/echo", attrRecord.path)
}

func TestGetRequestAttributesValidSubResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTP_HEADER_REQ_URI, "v1/function/func_name/foo/bar")
	request.Header.Add(HTTP_HEADER_ORIG_METHOD, "GET")
	request.Header.Add(HTTP_HEADER_FWD_EMAIL, "super-admin@example.com")
	attrRecord, _ := getRequestAttributes(request)
	assert.Equal(t, "super-admin@example.com", attrRecord.userEmail)
	assert.Equal(t, ActionGet, attrRecord.action)
	assert.Equal(t, "function", attrRecord.resource)
	assert.Equal(t, true, attrRecord.isResourceRequest)
	assert.Equal(t, "", attrRecord.path)
}

func TestRedirectHandler(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	handlers := &Handlers{}
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	request := httptest.NewRequest("GET", "/v1/iam/redirect", nil)
	testCookie := &http.Cookie{
		Name:  "_oauth2_proxy",
		Value: "testCookie",
	}
	request.AddCookie(testCookie)
	params := operations.RedirectParams{
		HTTPRequest: request,
		Redirect:    swag.String("http://redirect.com"),
	}
	responder := api.RedirectHandler.Handle(params, nil)

	w := httptest.NewRecorder()
	responder.WriteResponse(w, runtime.JSONProducer())
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)

	location, err := resp.Location()
	assert.Nil(t, err)

	expectedCookie := url.Values{
		"cookie": {testCookie.String()},
	}
	assert.Equal(t, fmt.Sprintf("http://redirect.com?%s", expectedCookie.Encode()), location.String())
}
