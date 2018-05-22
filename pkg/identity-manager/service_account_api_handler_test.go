///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	serviceaccountOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/serviceaccount"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func newServiceAccountModel(name string, pubKeyEncoded string) *v1.ServiceAccount {

	return &v1.ServiceAccount{
		Name:      swag.String(name),
		PublicKey: &pubKeyEncoded,
	}
}

func setupServiceAccountTestAPI(t *testing.T) *operations.IdentityManagerAPI {
	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	pubKey, _ := ioutil.ReadFile("testdata/test_key.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name: "test-serviceaccount-1",
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	return api
}

func TestAddServiceAccountHandler(t *testing.T) {
	pubKey, _ := ioutil.ReadFile("testdata/test_key.pub")
	pubKeyEncoded := base64.StdEncoding.EncodeToString(pubKey)
	reqBody := newServiceAccountModel("test-serviceaccount-2", pubKeyEncoded)
	r := httptest.NewRequest("POST", "/v1/iam/serviceaccount", nil)
	params := serviceaccountOperations.AddServiceAccountParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountAddServiceAccountHandler.Handle(params, "testCookie")
	var respBody v1.ServiceAccount
	helpers.HandlerRequest(t, responder, &respBody, http.StatusCreated)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-serviceaccount-2", *respBody.Name)
	assert.Equal(t, pubKeyEncoded, *respBody.PublicKey)
}

func TestGetServiceAccountsHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/iam/serviceaccount", nil)
	params := serviceaccountOperations.GetServiceAccountsParams{
		HTTPRequest: r,
	}
	// Also, load test data
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountGetServiceAccountsHandler.Handle(params, "testCookie")
	var respBody []v1.ServiceAccount
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.Len(t, respBody, 1)
	assert.NotEmpty(t, respBody[0].ID)
	assert.NotNil(t, respBody[0].CreatedTime)
	assert.Equal(t, "test-serviceaccount-1", *respBody[0].Name)
}

func TestDeleteServiceAccountHandler(t *testing.T) {
	r := httptest.NewRequest("DELETE", "/v1/iam/serviceaccount/test-serviceaccount-1", nil)
	params := serviceaccountOperations.DeleteServiceAccountParams{
		HTTPRequest:        r,
		ServiceAccountName: "test-serviceaccount-1",
	}
	// Also, load test data
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountDeleteServiceAccountHandler.Handle(params, "testCookie")
	var respBody v1.ServiceAccount
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-serviceaccount-1", *respBody.Name)
	assert.Equal(t, "DELETING", string(respBody.Status))
}

func TestGetServiceAccountHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/iam/serviceaccount/test-serviceaccount-1", nil)
	params := serviceaccountOperations.GetServiceAccountParams{
		HTTPRequest:        r,
		ServiceAccountName: "test-serviceaccount-1",
	}
	// Also, load test data
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountGetServiceAccountHandler.Handle(params, "testCookie")
	var respBody v1.ServiceAccount
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)

	assert.NotEmpty(t, respBody.ID)
	assert.NotNil(t, respBody.CreatedTime)
	assert.Equal(t, "test-serviceaccount-1", *respBody.Name)
}

func TestGetServiceAccountHandlerNotFound(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/iam/serviceaccount/not-found-serviceaccount", nil)
	params := serviceaccountOperations.GetServiceAccountParams{
		HTTPRequest:        r,
		ServiceAccountName: "not-found-serviceaccount",
	}
	// Also, load test data
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountGetServiceAccountHandler.Handle(params, "testCookie")
	var respBody v1.ServiceAccount
	helpers.HandlerRequest(t, responder, &respBody, http.StatusNotFound)
}

func TestUpdateServiceAccountHandlerError(t *testing.T) {
	reqBody := newServiceAccountModel("test-serviceaccount-1", "^notbase64encoded")

	r := httptest.NewRequest("UPDATE", "/v1/iam/serviceaccount/test-serviceaccount-1", nil)
	params := serviceaccountOperations.UpdateServiceAccountParams{
		HTTPRequest:        r,
		ServiceAccountName: "test-serviceaccount-1",
		Body:               reqBody,
	}

	// Also, load test data
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountUpdateServiceAccountHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, http.StatusBadRequest)
	assert.Equal(t, "error validating service account: public key is not base64 encoded", *respBody.Message)
}

func TestUpdateServiceAccountHandlerInvalidKey(t *testing.T) {
	reqBody := newServiceAccountModel("test-serviceaccount-1", "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0=")

	r := httptest.NewRequest("UPDATE", "/v1/iam/serviceaccount/test-serviceaccount-1", nil)
	params := serviceaccountOperations.UpdateServiceAccountParams{
		HTTPRequest:        r,
		ServiceAccountName: "test-serviceaccount-1",
		Body:               reqBody,
	}

	// Also, load test data
	api := setupServiceAccountTestAPI(t)
	responder := api.ServiceaccountUpdateServiceAccountHandler.Handle(params, "testCookie")
	var respBody v1.Error
	helpers.HandlerRequest(t, responder, &respBody, http.StatusBadRequest)
	assert.Equal(t, "error validating service account: invalid public key or public key not in PEM format", *respBody.Message)
}
