///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-openapi/swag"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	organizationOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/organization"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func newOrganizationModel(name string) *v1.Organization {
	return &v1.Organization{
		Name: swag.String(name),
	}
}

func setupOrgTestAPI(t *testing.T) *operations.IdentityManagerAPI {
	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	org := &Organization{
		BaseEntity: entitystore.BaseEntity{
			Name: "test-organization-1",
		},
	}
	es.Add(context.Background(), org)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	return api
}

func TestAddOrganizationHandler(t *testing.T) {

	reqBody := newOrganizationModel("test-organization-2")
	r := httptest.NewRequest("POST", "/v1/iam/organization", nil)
	params := organizationOperations.AddOrganizationParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	api := setupOrgTestAPI(t)
	responder := api.OrganizationAddOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusCreated)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-organization-2", *respBody.Name)
}

func TestAddOrganizationHandlerDuplicateOrganization(t *testing.T) {

	reqBody := newOrganizationModel("test-organization-1")
	r := httptest.NewRequest("POST", "/v1/iam/organization", nil)
	params := organizationOperations.AddOrganizationParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	// Pre-create organization with same name
	api := setupOrgTestAPI(t)
	responder := api.OrganizationAddOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, http.StatusConflict)
	assert.EqualValues(t, http.StatusConflict, respBody.Code)
}

func TestGetOrganizationsHandler(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/organization", nil)
	params := organizationOperations.GetOrganizationsParams{
		HTTPRequest: r,
	}
	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationGetOrganizationsHandler.Handle(params, "testCookie")
	var respBody []v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.Len(t, respBody, 1)
	assert.NotEmpty(t, respBody[0].ID)
	assert.NotNil(t, respBody[0].CreatedTime)
	assert.Equal(t, "test-organization-1", *respBody[0].Name)
}

func TestDeleteOrganizationHandler(t *testing.T) {

	r := httptest.NewRequest("DELETE", "/v1/iam/organization/test-organization-1", nil)
	params := organizationOperations.DeleteOrganizationParams{
		HTTPRequest:      r,
		OrganizationName: "test-organization-1",
	}
	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationDeleteOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-organization-1", *respBody.Name)
	assert.Equal(t, v1.StatusDELETING, respBody.Status)

	// Try, deleting again - Bad request
	responder = api.OrganizationDeleteOrganizationHandler.Handle(params, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, http.StatusBadRequest)
}

func TestDeleteOrganizationHandlerNotFound(t *testing.T) {

	r := httptest.NewRequest("DELETE", "/v1/iam/organization/test-organization-unknown", nil)
	params := organizationOperations.DeleteOrganizationParams{
		HTTPRequest:      r,
		OrganizationName: "test-organization-unknown",
	}
	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationDeleteOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)

}

func TestGetOrganizationHandler(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/organization/test-organization-1", nil)
	params := organizationOperations.GetOrganizationParams{
		HTTPRequest:      r,
		OrganizationName: "test-organization-1",
	}
	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationGetOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-organization-1", *respBody.Name)
}

func TestGetOrganizationHandlerNotFound(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/iam/organization/not-found-organization", nil)
	params := organizationOperations.GetOrganizationParams{
		HTTPRequest:      r,
		OrganizationName: "not-found-organization",
	}
	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationGetOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}

func TestUpdateOrganizationHandler(t *testing.T) {

	reqBody := newOrganizationModel("test-organization-1")

	r := httptest.NewRequest("UPDATE", "/v1/iam/organization/test-organization-1", nil)
	params := organizationOperations.UpdateOrganizationParams{
		HTTPRequest:      r,
		OrganizationName: "test-organization-1",
		Body:             reqBody,
	}

	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationUpdateOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-organization-1", *respBody.Name)
}

func TestUpdateOrganizationHandlerNotFound(t *testing.T) {

	reqBody := newOrganizationModel("not-found-organization")

	r := httptest.NewRequest("UPDATE", "/v1/iam/organization/not-found-organization", nil)
	params := organizationOperations.UpdateOrganizationParams{
		HTTPRequest:      r,
		OrganizationName: "not-found-organization",
		Body:             reqBody,
	}

	// Also, load test data
	api := setupOrgTestAPI(t)
	responder := api.OrganizationUpdateOrganizationHandler.Handle(params, "testCookie")
	var respBody v1.Organization
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}
