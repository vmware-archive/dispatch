///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-openapi/swag"
	"github.com/vmware/dispatch/pkg/api/v1"
	policyOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/policy"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func newPolicyModel(name string, subjects []string, resources []string, actions []string) *v1.Policy {
	return &v1.Policy{
		Name: swag.String(name),
		Rules: []*v1.Rule{
			{
				Subjects:  subjects,
				Resources: resources,
				Actions:   actions,
			},
		},
	}
}

func TestAddPolicyHandler(t *testing.T) {

	subjects := []string{"user@example.com"}
	resources := []string{"*"}
	actions := []string{"get"}

	reqBody := newPolicyModel("test-policy-1", subjects, resources, actions)
	r := httptest.NewRequest("POST", "/v1/iam/policy", nil)
	params := policyOperations.AddPolicyParams{
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgA,
	}
	api := setupTestAPI(t, false)
	responder := api.PolicyAddPolicyHandler.Handle(params, "testCookie")
	var respBody v1.Policy
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
	var respBody v1.Error
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
		HTTPRequest:  r,
		Body:         reqBody,
		XDispatchOrg: testOrgA,
	}
	// Pre-create policy with same name
	api := setupTestAPI(t, true)
	responder := api.PolicyAddPolicyHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, http.StatusConflict)
	assert.EqualValues(t, http.StatusConflict, respBody.Code)
}

func TestGetPoliciesHandler(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/policy", nil)
	params := policyOperations.GetPoliciesParams{
		HTTPRequest:  r,
		XDispatchOrg: testOrgA,
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyGetPoliciesHandler.Handle(params, "testCookie")
	var respBody []v1.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.Len(t, respBody, 2)
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
		HTTPRequest:  r,
		PolicyName:   "test-policy-1",
		XDispatchOrg: testOrgA,
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyDeletePolicyHandler.Handle(params, "testCookie")
	var respBody v1.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-policy-1", *respBody.Name)
	assert.Equal(t, []string{"readonly-user@example.com"}, respBody.Rules[0].Subjects)
	assert.Equal(t, []string{"*"}, respBody.Rules[0].Resources)
	assert.Equal(t, []string{"get"}, respBody.Rules[0].Actions)
	assert.Equal(t, v1.StatusDELETING, respBody.Status)

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
	var respBody v1.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)

}

func TestGetPolicyHandler(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/policy/test-policy-1", nil)
	params := policyOperations.GetPolicyParams{
		HTTPRequest:  r,
		PolicyName:   "test-policy-1",
		XDispatchOrg: testOrgA,
	}
	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyGetPolicyHandler.Handle(params, "testCookie")
	var respBody v1.Policy
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
	var respBody v1.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}

func TestUpdatePolicyHandler(t *testing.T) {

	subjects := []string{"user2@example.com"}
	resources := []string{"*"}
	actions := []string{"delete"}

	reqBody := newPolicyModel("test-policy-1", subjects, resources, actions)

	r := httptest.NewRequest("UPDATE", "/v1/iam/policy/test-policy-1", nil)
	params := policyOperations.UpdatePolicyParams{
		HTTPRequest:  r,
		PolicyName:   "test-policy-1",
		Body:         reqBody,
		XDispatchOrg: testOrgA,
	}

	// Also, load test data
	api := setupTestAPI(t, true)
	responder := api.PolicyUpdatePolicyHandler.Handle(params, "testCookie")
	var respBody v1.Policy
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
	var respBody v1.Policy
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}
