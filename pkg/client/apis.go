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
	"github.com/go-openapi/swag"
	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/endpoints/gen/client"
	"github.com/vmware/dispatch/pkg/endpoints/gen/client/endpoint"
)

// EndpointsClient defines the api client interface
type EndpointsClient interface {
	// Endpoints
	CreateEndpoint(ctx context.Context, organizationID string, model *v1.Endpoint) (*v1.Endpoint, error)
	DeleteEndpoint(ctx context.Context, organizationID string, name string) (*v1.Endpoint, error)
	UpdateEndpoint(ctx context.Context, organizationID string, model *v1.Endpoint) (*v1.Endpoint, error)
	GetEndpoint(ctx context.Context, organizationID string, name string) (*v1.Endpoint, error)
	ListEndpoints(ctx context.Context, organizationID string) ([]v1.Endpoint, error)
}

// NewEndpointsClient is used to create a new Endpoints client
func NewEndpointsClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) *DefaultEndpointsClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultEndpointsClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultEndpointsClient defines the default Endpoints client
type DefaultEndpointsClient struct {
	baseClient

	client *swaggerclient.Endpoints
	auth   runtime.ClientAuthInfoWriter
}

// CreateEndpoint creates new api
func (c *DefaultEndpointsClient) CreateEndpoint(ctx context.Context, organizationID string, model *v1.Endpoint) (*v1.Endpoint, error) {
	params := endpoint.AddEndpointParams{
		Context:      ctx,
		Body:         model,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Endpoint.AddEndpoint(&params, c.auth)
	if err != nil {
		return nil, createEndpointSwaggerError(err)
	}
	return response.Payload, nil
}

func createEndpointSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.AddEndpointBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.AddEndpointUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.AddEndpointForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.AddEndpointConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *endpoint.AddEndpointDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteEndpoint deletes an api
func (c *DefaultEndpointsClient) DeleteEndpoint(ctx context.Context, organizationID string, name string) (*v1.Endpoint, error) {
	params := endpoint.DeleteEndpointParams{
		Context:      ctx,
		Endpoint:     name,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Endpoint.DeleteEndpoint(&params, c.auth)
	if err != nil {
		return nil, deleteEndpointSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteEndpointSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.DeleteEndpointBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.DeleteEndpointUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.DeleteEndpointForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.DeleteEndpointNotFound:
		return NewErrorNotFound(v.Payload)
	case *endpoint.DeleteEndpointDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateEndpoint updates an api
func (c *DefaultEndpointsClient) UpdateEndpoint(ctx context.Context, organizationID string, model *v1.Endpoint) (*v1.Endpoint, error) {
	params := endpoint.UpdateEndpointParams{
		Context:      ctx,
		Body:         model,
		Endpoint:     model.Name,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Endpoint.UpdateEndpoint(&params, c.auth)
	if err != nil {
		return nil, updateEndpointSwaggerError(err)
	}
	return response.Payload, nil
}

func updateEndpointSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.UpdateEndpointBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.UpdateEndpointUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.UpdateEndpointForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.UpdateEndpointNotFound:
		return NewErrorNotFound(v.Payload)
	case *endpoint.UpdateEndpointDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetEndpoint retrieves an api
func (c *DefaultEndpointsClient) GetEndpoint(ctx context.Context, organizationID string, name string) (*v1.Endpoint, error) {
	params := endpoint.GetEndpointParams{
		Context:      ctx,
		Endpoint:     name,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Endpoint.GetEndpoint(&params, c.auth)
	if err != nil {
		return nil, getEndpointSwaggerError(err)
	}
	return response.Payload, nil
}

func getEndpointSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.GetEndpointBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *endpoint.GetEndpointUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.GetEndpointForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.GetEndpointNotFound:
		return NewErrorNotFound(v.Payload)
	case *endpoint.GetEndpointDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListEndpoints returns a list of Endpoints
func (c *DefaultEndpointsClient) ListEndpoints(ctx context.Context, organizationID string) ([]v1.Endpoint, error) {
	params := endpoint.GetEndpointsParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Endpoint.GetEndpoints(&params, c.auth)
	if err != nil {
		return nil, listEndpointsSwaggerError(err)
	}

	apis := []v1.Endpoint{}
	for _, api := range response.Payload {
		apis = append(apis, *api)
	}
	return apis, nil
}

func listEndpointsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *endpoint.GetEndpointsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *endpoint.GetEndpointsForbidden:
		return NewErrorForbidden(v.Payload)
	case *endpoint.GetEndpointsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
