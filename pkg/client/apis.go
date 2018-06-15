///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	swaggerclient "github.com/vmware/dispatch/pkg/api-manager/gen/client"
	"github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/api/v1"
)

// APIsClient defines the api client interface
type APIsClient interface {
	// APIs
	CreateAPI(ctx context.Context, organizationID string, api *v1.API) (*v1.API, error)
	DeleteAPI(ctx context.Context, organizationID string, apiName string) (*v1.API, error)
	UpdateAPI(ctx context.Context, organizationID string, api *v1.API) (*v1.API, error)
	GetAPI(ctx context.Context, organizationID string, apiName string) (*v1.API, error)
	ListAPIs(ctx context.Context, organizationID string) ([]v1.API, error)
}

// NewAPIsClient is used to create a new APIs client
func NewAPIsClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) *DefaultAPIsClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultAPIsClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultAPIsClient defines the default APIs client
type DefaultAPIsClient struct {
	baseClient

	client *swaggerclient.APIManager
	auth   runtime.ClientAuthInfoWriter
}

// CreateAPI creates new api
func (c *DefaultAPIsClient) CreateAPI(ctx context.Context, organizationID string, api *v1.API) (*v1.API, error) {
	params := endpoint.AddAPIParams{
		Context:      ctx,
		Body:         api,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Endpoint.AddAPI(&params, c.auth)
	if err != nil {
		return nil, createAPISwaggerError(err)
	}
	return response.Payload, nil
}

func createAPISwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.AddAPIBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.AddAPIUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.AddAPIForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.AddAPIConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *endpoint.AddAPIDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteAPI deletes an api
func (c *DefaultAPIsClient) DeleteAPI(ctx context.Context, organizationID string, apiName string) (*v1.API, error) {
	params := endpoint.DeleteAPIParams{
		Context:      ctx,
		API:          apiName,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Endpoint.DeleteAPI(&params, c.auth)
	if err != nil {
		return nil, deleteAPISwaggerError(err)
	}
	return response.Payload, nil
}

func deleteAPISwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.DeleteAPIBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.DeleteAPIUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.DeleteAPIForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.DeleteAPINotFound:
		return NewErrorNotFound(v.Payload)
	case *endpoint.DeleteAPIDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateAPI updates an api
func (c *DefaultAPIsClient) UpdateAPI(ctx context.Context, organizationID string, api *v1.API) (*v1.API, error) {
	params := endpoint.UpdateAPIParams{
		Context:      ctx,
		Body:         api,
		API:          *api.Name,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Endpoint.UpdateAPI(&params, c.auth)
	if err != nil {
		return nil, updateAPISwaggerError(err)
	}
	return response.Payload, nil
}

func updateAPISwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.UpdateAPIBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.UpdateAPIUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.UpdateAPIForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.UpdateAPINotFound:
		return NewErrorNotFound(v.Payload)
	case *endpoint.UpdateAPIDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetAPI retrieves an api
func (c *DefaultAPIsClient) GetAPI(ctx context.Context, organizationID string, apiName string) (*v1.API, error) {
	params := endpoint.GetAPIParams{
		Context:      ctx,
		API:          apiName,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Endpoint.GetAPI(&params, c.auth)
	if err != nil {
		return nil, getAPISwaggerError(err)
	}
	return response.Payload, nil
}

func getAPISwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.GetAPIBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.GetAPIUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.GetAPIForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.GetAPINotFound:
		return NewErrorNotFound(v.Payload)
	case *endpoint.GetAPIDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListAPIs returns a list of APIs
func (c *DefaultAPIsClient) ListAPIs(ctx context.Context, organizationID string) ([]v1.API, error) {
	params := endpoint.GetApisParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Endpoint.GetApis(&params, c.auth)
	if err != nil {
		return nil, listAPIsSwaggerError(err)
	}

	apis := []v1.API{}
	for _, api := range response.Payload {
		apis = append(apis, *api)
	}
	return apis, nil
}

func listAPIsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.GetApisUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.GetApisForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.GetApisDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
