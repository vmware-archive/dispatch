///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"encoding/json"
	"fmt"
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

func TestCookieAuth(t *testing.T) {
	authService := GetTestAuthService(t)
	testIDToken := &gooidc.IDToken{
		Subject: "testUserCookie",
	}
	testSessionID, err := authService.CreateAndSaveSession(testIDToken)
	assert.NoError(t, err, "Unexpected Error")
	api := MakeAPI(t, authService)

	testToken := fmt.Sprintf("sessionId=%s; Path=/", testSessionID)

	session, err := api.CookieAuth(testToken)
	assert.NoError(t, err, "Unexpected Error")
	assert.IsType(t, &models.Session{}, session, "Unexpected Type Error")
	sessionModel := session.(*models.Session)

	assert.Equal(t, *sessionModel.Name, "testUserCookie")
	assert.Equal(t, *sessionModel.Name, testSessionID)

	testTokenInvalid := "sessionId=wrong; Path=/"
	session, err = api.CookieAuth(testTokenInvalid)
	assert.Error(t, err, "Cookie Auth Should Fail")
	assert.Nil(t, session)

	err = authService.RemoveSession(testSessionID)
	assert.NoError(t, err, "Unexpected Error")
}

func TestAuthenticationLoginHandler(t *testing.T) {

	authService := GetTestAuthService(t)
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

	authService := GetTestAuthService(t)
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

	sessionId, err := ParseDefaultCookie(resp.Header.Get("Set-Cookie"))
	assert.NoError(t, err, "Unexpected Error")

	session, err := authService.GetSession(sessionId)
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, "testUser1", session.Name)
}

func TestAuthenticationLogoutHandler(t *testing.T) {

	authService := GetTestAuthService(t)
	testIDToken1 := &gooidc.IDToken{
		Subject: "testUser1",
	}
	testSessionID1, err := authService.CreateAndSaveSession(testIDToken1)
	assert.NoError(t, err, "Unexpected Error")
	api := MakeAPI(t, authService)

	req := httptest.NewRequest("GET", "/v1/iam/logout", nil)
	params := authentication.LogoutParams{HTTPRequest: req}
	session := &models.Session{
		Name: swag.String(testIDToken1.Subject),
		ID:   swag.String(testSessionID1),
	}
	responder := api.AuthenticationLogoutHandler.Handle(params, session)

	var realRespBody models.Message
	resp := HandlerRequest(t, responder, &realRespBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "You Have Successfully Logged Out", *realRespBody.Message)

	err = authService.RemoveSession(testSessionID1)
	assert.Error(t, err, "Session should have been removed")
}

func TestHomeHandler(t *testing.T) {

	authService := GetTestAuthService(t)
	api := MakeAPI(t, authService)

	req := httptest.NewRequest("GET", "/v1/iam/home", nil)
	params := operations.HomeParams{HTTPRequest: req}

	responder := api.HomeHandler.Handle(params)

	var realRespBody models.Message
	resp := HandlerRequest(t, responder, &realRespBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Home Page, You have already logged in", *realRespBody.Message)
}

// MakeAPI returns an API for testing
func MakeAPI(t *testing.T, authService *AuthService) *operations.IdentityManagerAPI {

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewIdentityManagerAPI(swaggerSpec)
	handlers := &Handlers{authService}
	handlers.ConfigureHandlers(api)
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
