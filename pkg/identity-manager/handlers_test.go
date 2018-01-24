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
	"path/filepath"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"

	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestHomeHandler(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	handlers := &Handlers{}
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
	handlers := &Handlers{}
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
	handlers := &Handlers{}
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	IdentityManagerFlags.PolicyFile = filepath.Join("testdata", "test_policy.csv")
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add("X-ORIGINAL-METHOD", "GET")
	request.Header.Add("X-FORWARDED-EMAIL", "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerPolicyFail(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	handlers := &Handlers{}
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	IdentityManagerFlags.PolicyFile = filepath.Join("testdata", "test_policy.csv")
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add("X-ORIGINAL-METHOD", "POST")
	request.Header.Add("X-FORWARDED-EMAIL", "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerWithoutPolicyFile(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	handlers := &Handlers{}
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	request := httptest.NewRequest("GET", "/auth", nil)
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyNoHeaders(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	handlers := &Handlers{}
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	IdentityManagerFlags.PolicyFile = filepath.Join("testdata", "test_policy.csv")
	request := httptest.NewRequest("GET", "/auth", nil)
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestGetRequestAttributes(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add("X-ORIGINAL-METHOD", "POST")
	request.Header.Add("X-FORWARDED-EMAIL", "super-admin@example.com")
	attrRecord := getRequestAttributes(request)
	assert.Equal(t, "super-admin@example.com", attrRecord.userEmail)
	assert.Equal(t, ActionCreate, attrRecord.action)
	assert.Equal(t, "*", attrRecord.resource)
}

func TestGetRequestAttributesNoHeaders(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	attrRecord := getRequestAttributes(request)
	assert.Equal(t, "", attrRecord.userEmail)
	assert.Equal(t, attrRecord.action, attrRecord.action)
	assert.Equal(t, "*", attrRecord.resource)
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
