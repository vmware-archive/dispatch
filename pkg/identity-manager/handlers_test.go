///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	gooidc "github.com/coreos/go-oidc"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi/operations"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi/operations/authentication"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/mocks"
)

func TestAuthenticationLoginHandler(t *testing.T) {

	authService := NewAuthService(TestConfig)
	mockedOIDC := new(mocks.OIDC)
	mockedOIDC.On("GetAuthEndpoint", "foobar").Return("example.com/oidc")
	authService.Oidc = mockedOIDC

	api := MakeAPI(t, authService)

	req := httptest.NewRequest("GET", "/v1/iam/", nil)
	params := authentication.LoginParams{HTTPRequest: req}

	responder := api.AuthenticationLoginHandler.Handle(params)

	resp := HandlerRequest(t, responder, nil)
	assert.Equal(t, "example.com/oidc", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestAuthenticationLoginVmwareHandler(t *testing.T) {

	authService := NewAuthService(TestConfig)
	mockedOIDC := new(mocks.OIDC)
	mockedOIDC.On("ExchangeIDToken", "exampleCode").Return(
		&gooidc.IDToken{
			Issuer:   "FakedIssuer",
			Audience: []string{"aud1", "aud2"},
			Subject:  "testUser1"}, nil)
	authService.Oidc = mockedOIDC

	api := MakeAPI(t, authService)

	req := httptest.NewRequest("GET", "/v1/iam/login/vmware", nil)
	params := authentication.LoginVmwareParams{
		HTTPRequest: req,
		Code:        swag.String("exampleCode"),
		State:       swag.String("foobar")}
	responder := api.AuthenticationLoginVmwareHandler.Handle(params)

	resp := HandlerRequest(t, responder, nil)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/v1/iam/home", resp.Header.Get("Location"))
	assert.Equal(t, "username=testUser1; Path=/", resp.Header.Get("Set-Cookie"))
}

func TestAuthenticationLogoutHandler(t *testing.T) {

	authService := NewAuthService(TestConfig)
	authService.CookieStore.SaveCookie("testUser1", "testUser1Value")
	api := MakeAPI(t, authService)

	req := httptest.NewRequest("GET", "/v1/iam/logout", nil)
	params := authentication.LogoutParams{HTTPRequest: req}
	principle := &models.Principle{Name: swag.String("testUser1")}

	responder := api.AuthenticationLogoutHandler.Handle(params, principle)

	var realRespBody models.Message
	resp := HandlerRequest(t, responder, &realRespBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "You Have Successfully Logged Out", *realRespBody.Message)
}

func TestHomeHandler(t *testing.T) {
	authService := NewAuthService(TestConfig)
	authService.CookieStore.SaveCookie("testUser1", "testUser1Value")
	api := MakeAPI(t, authService)

	req := httptest.NewRequest("GET", "/v1/iam/home", nil)
	params := operations.HomeParams{HTTPRequest: req}
	principle := &models.Principle{Name: swag.String("testUser1")}

	responder := api.HomeHandler.Handle(params, principle)

	var realRespBody models.Message
	resp := HandlerRequest(t, responder, &realRespBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Hello testUser1", *realRespBody.Message)
}

// MakeAPI returns an API for testing
func MakeAPI(t *testing.T, authService *AuthService) *operations.IdentityManagerAPI {

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewIdentityManagerAPI(swaggerSpec)
	ConfigureHandlers(api, authService)
	return api
}

// HandlerRequest is a convenience function for testing API handlers
func HandlerRequest(t *testing.T, responder middleware.Responder, responseObject interface{}) *http.Response {
	w := httptest.NewRecorder()
	responder.WriteResponse(w, runtime.JSONProducer())
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading back response: %v", err)
	}
	if responseObject != nil {
		err = json.Unmarshal(body, responseObject)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
	}
	return resp
}
