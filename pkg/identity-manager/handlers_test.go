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
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	policyOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/policy"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func addTestData(store entitystore.EntityStore) {
	// Add test policies and rules
	e := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           "test-policy-1",
			Status:         entitystore.StatusREADY,
		},
		Rules: []Rule{
			{
				Subjects:  []string{"readonly-user@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"get"},
			},
			{
				Subjects:  []string{"super-admin@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			}},
	}
	store.Add(e)
}

func setupTestAPI(t *testing.T, addTestPolicies bool) *operations.IdentityManagerAPI {
	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	if addTestPolicies {
		addTestData(es)
	}
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	return api
}

func newPolicyModel(name string, subjects []string, resources []string, actions []string) *models.Policy {
	return &models.Policy{
		Name: swag.String(name),
		Rules: models.PolicyRules{
			&models.Rule{
				Subjects:  subjects,
				Resources: resources,
				Actions:   actions,
			},
		},
	}
}

func TestHomeHandler(t *testing.T) {
	api := setupTestAPI(t, false)

	params := operations.HomeParams{
		HTTPRequest: httptest.NewRequest("GET", "/v1/iam/home", nil),
	}
	responder := api.HomeHandler.Handle(params, nil)

	var respBody models.Message
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)
	assert.Equal(t, "Home Page, You have already logged in", *respBody.Message)
}

func TestRootHandler(t *testing.T) {

	api := setupTestAPI(t, false)

	params := operations.RootParams{
		HTTPRequest: httptest.NewRequest("GET", "/", nil),
	}
	responder := api.RootHandler.Handle(params)

	var respBody models.Message
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)
	assert.Equal(t, "Default Root Page", *respBody.Message)
}

func TestAuthHandlerPolicyPass(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerWithoutPolicyData(t *testing.T) {

	api := setupTestAPI(t, false)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerNonResourcePass(t *testing.T) {

	api := setupTestAPI(t, false)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/echo")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "noname@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerBootstrapFail(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "bootstrap-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode but don't set user
	IdentityManagerFlags.EnableBootstrapMode = true
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerBootstrapPass(t *testing.T) {

	//bootstrap user can only access iam resource
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/iam/policy")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "bootstrap-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode and user
	IdentityManagerFlags.EnableBootstrapMode = true
	os.Setenv("IAM_BOOTSTRAP_USER", "bootstrap-user@example.com")
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerBootstrapForbid(t *testing.T) {

	//bootstrap user can only access iam resource, will get forbidden when accessing other resources
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "bootstrap-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode and user
	IdentityManagerFlags.EnableBootstrapMode = true
	os.Setenv("IAM_BOOTSTRAP_USER", "bootstrap-user@example.com")
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerNonBootstrapUserInBootstrapMode(t *testing.T) {

	//non-bootstrap user in bootstrap mode cannot access any resources
	api := setupTestAPI(t, true)

	// try access iam resources
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/iam/policy")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "non-bootstrap-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode and user
	IdentityManagerFlags.EnableBootstrapMode = true
	os.Setenv("IAM_BOOTSTRAP_USER", "bootstrap-user@example.com")
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)

	// try access non-iam resources
	request.Header.Set(HTTPHeaderReqURI, "v1/image")
	responder = api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyFail(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	request.Header.Add(HTTPHeaderFwdEmail, "readonly-user@example.com")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyNoValidHeader(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	// Missing Email Header
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, nil)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestGetRequestAttributesNoEmailHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	_, err := getRequestAttributes(request)
	assert.EqualError(t, err, "X-Forwarded-Email header not found")
}

func TestGetRequestAttributesNoMethodHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderFwdEmail, "super-admin@example.com")
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	_, err := getRequestAttributes(request)
	assert.EqualError(t, err, "X-Original-Method header not found")
}

func TestGetRequestAttributesNoURLHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderFwdEmail, "super-admin@example.com")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	_, err := getRequestAttributes(request)
	assert.EqualError(t, err, "X-Auth-Request-Redirect header not found")
}

func TestGetRequestAttributesValidResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	request.Header.Add(HTTPHeaderFwdEmail, "super-admin@example.com")
	attrRecord, _ := getRequestAttributes(request)
	assert.Equal(t, "super-admin@example.com", attrRecord.userEmail)
	assert.Equal(t, ActionCreate, attrRecord.action)
	assert.Equal(t, "function", attrRecord.resource)
	assert.Equal(t, true, attrRecord.isResourceRequest)
	assert.Equal(t, "", attrRecord.path)
}

func TestGetRequestAttributesNonResourcePath(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/echo")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "super-admin@example.com")
	attrRecord, _ := getRequestAttributes(request)
	assert.Equal(t, "super-admin@example.com", attrRecord.userEmail)
	assert.Equal(t, ActionGet, attrRecord.action)
	assert.Equal(t, "", attrRecord.resource)
	assert.Equal(t, false, attrRecord.isResourceRequest)
	assert.Equal(t, "/echo", attrRecord.path)
}

func TestGetRequestAttributesValidSubResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "v1/function/func_name/foo/bar")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	request.Header.Add(HTTPHeaderFwdEmail, "super-admin@example.com")
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

func TestAddPolicyHandler(t *testing.T) {

	subjects := []string{"user@example.com"}
	resources := []string{"*"}
	actions := []string{"get"}

	reqBody := newPolicyModel("test-policy-1", subjects, resources, actions)
	r := httptest.NewRequest("POST", "/v1/iam/policy", nil)
	params := policyOperations.AddPolicyParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	api := setupTestAPI(t, false)
	responder := api.PolicyAddPolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusCreated)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-policy-1", *respBody.Name)
	assert.Equal(t, subjects, respBody.Rules[0].Subjects)
	assert.Equal(t, resources, respBody.Rules[0].Resources)
	assert.Equal(t, actions, respBody.Rules[0].Actions)
}

func TestAddPolicyHandlerBasicValidation(t *testing.T) {

	subjects := []string{"user@example.com"}

	reqBody := newPolicyModel("test-policy-1", subjects, nil, nil)
	r := httptest.NewRequest("POST", "/v1/iam/policy", nil)
	params := policyOperations.AddPolicyParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	api := setupTestAPI(t, false)
	responder := api.PolicyAddPolicyHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, http.StatusBadRequest)
	assert.EqualValues(t, http.StatusBadRequest, respBody.Code)
}

func TestAddPolicyHandlerDuplicatePolicy(t *testing.T) {

	subjects := []string{"user@example.com"}
	resources := []string{"*"}
	actions := []string{"get"}

	reqBody := newPolicyModel("test-policy-1", subjects, resources, actions)
	r := httptest.NewRequest("POST", "/v1/iam/policy", nil)
	params := policyOperations.AddPolicyParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	// Pre-create policy with same name
	api := setupTestAPI(t, true)
	responder := api.PolicyAddPolicyHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, http.StatusConflict)
	assert.EqualValues(t, http.StatusConflict, respBody.Code)
}

func TestGetPoliciesHandler(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/policy", nil)
	params := policyOperations.GetPoliciesParams{
		HTTPRequest: r,
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyGetPoliciesHandler.Handle(params, "testCookie")
	var respBody []models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.Len(t, respBody, 1)
	assert.NotEmpty(t, respBody[0].ID)
	assert.NotNil(t, respBody[0].CreatedTime)
	assert.Equal(t, "test-policy-1", *respBody[0].Name)
	assert.Equal(t, []string{"readonly-user@example.com"}, respBody[0].Rules[0].Subjects)
	assert.Equal(t, []string{"*"}, respBody[0].Rules[0].Resources)
	assert.Equal(t, []string{"get"}, respBody[0].Rules[0].Actions)
}

func TestDeletePolicyHandler(t *testing.T) {

	r := httptest.NewRequest("DELETE", "/v1/iam/policy/test-policy-1", nil)
	params := policyOperations.DeletePolicyParams{
		HTTPRequest: r,
		PolicyName:  "test-policy-1",
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyDeletePolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-policy-1", *respBody.Name)
	assert.Equal(t, []string{"readonly-user@example.com"}, respBody.Rules[0].Subjects)
	assert.Equal(t, []string{"*"}, respBody.Rules[0].Resources)
	assert.Equal(t, []string{"get"}, respBody.Rules[0].Actions)
	assert.Equal(t, models.StatusDELETING, respBody.Status)

	// Try, deleting again - Bad request
	responder = api.PolicyDeletePolicyHandler.Handle(params, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, http.StatusBadRequest)
}

func TestDeletePolicyHandlerNotFound(t *testing.T) {

	r := httptest.NewRequest("DELETE", "/v1/iam/policy/test-policy-unknown", nil)
	params := policyOperations.DeletePolicyParams{
		HTTPRequest: r,
		PolicyName:  "test-policy-unknown",
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyDeletePolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)

}

func TestGetPolicyHandler(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/policy/test-policy-1", nil)
	params := policyOperations.GetPolicyParams{
		HTTPRequest: r,
		PolicyName:  "test-policy-1",
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyGetPolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-policy-1", *respBody.Name)
	assert.Equal(t, []string{"readonly-user@example.com"}, respBody.Rules[0].Subjects)
	assert.Equal(t, []string{"*"}, respBody.Rules[0].Resources)
	assert.Equal(t, []string{"get"}, respBody.Rules[0].Actions)
}

func TestGetPolicyHandlerNotFound(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/policy/not-found-policy", nil)
	params := policyOperations.GetPolicyParams{
		HTTPRequest: r,
		PolicyName:  "not-found-policy",
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyGetPolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}

func TestUpdatePolicyHandler(t *testing.T) {

	subjects := []string{"user2@example.com"}
	resources := []string{"*"}
	actions := []string{"delete"}

	reqBody := newPolicyModel("test-policy-1", subjects, resources, actions)

	r := httptest.NewRequest("UPDATE", "/v1/iam/policy/test-policy-1", nil)
	params := policyOperations.UpdatePolicyParams{
		HTTPRequest: r,
		PolicyName:  "test-policy-1",
		Body:        reqBody,
	}

	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyUpdatePolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-policy-1", *respBody.Name)
	assert.Equal(t, subjects, respBody.Rules[0].Subjects)
	assert.Equal(t, resources, respBody.Rules[0].Resources)
	assert.Equal(t, actions, respBody.Rules[0].Actions)
}

func TestUpdatePolicyHandlerNotFound(t *testing.T) {

	subjects := []string{"user@example.com"}
	resources := []string{"*"}
	actions := []string{"delete"}

	reqBody := newPolicyModel("not-found-policy", subjects, resources, actions)

	r := httptest.NewRequest("UPDATE", "/v1/iam/policy/not-found-policy", nil)
	params := policyOperations.UpdatePolicyParams{
		HTTPRequest: r,
		PolicyName:  "not-found-policy",
		Body:        reqBody,
	}

	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyUpdatePolicyHandler.Handle(params, "testCookie")
	var respBody models.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}
