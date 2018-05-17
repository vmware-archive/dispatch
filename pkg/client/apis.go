///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "error when creating the api")
	}
	return response.Payload, nil
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
		return nil, errors.Wrap(err, "error when deleting the api")
	}
	return response.Payload, nil
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
		return nil, errors.Wrap(err, "error when updating the api")
	}
	return response.Payload, nil
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
		return nil, errors.Wrap(err, "error when getting the api")
	}
	return response.Payload, nil
}

// ListAPIs returns a list of APIs
func (c *DefaultAPIsClient) ListAPIs(ctx context.Context, organizationID string) ([]v1.API, error) {
	params := endpoint.GetApisParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Endpoint.GetApis(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when listing apis")
	}

	apis := []v1.API{}
	for _, api := range response.Payload {
		apis = append(apis, *api)
	}
	return apis, nil
}
