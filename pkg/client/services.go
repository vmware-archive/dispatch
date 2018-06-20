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
	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/service-manager/gen/client"
	serviceclassclient "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_class"
	serviceinstanceclient "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"
)

// ServicesClient defines the services client interface
type ServicesClient interface {
	// Service Instances
	CreateServiceInstance(ctx context.Context, organizationID string, serviceInstance *v1.ServiceInstance) (*v1.ServiceInstance, error)
	DeleteServiceInstance(ctx context.Context, organizationID string, serviceInstanceName string) error
	GetServiceInstance(ctx context.Context, organizationID string, serviceInstanceName string) (*v1.ServiceInstance, error)
	ListServiceInstances(ctx context.Context, organizationID string) ([]v1.ServiceInstance, error)

	// Service Classes
	GetServiceClass(ctx context.Context, organizationID string, serviceClassName string) (*v1.ServiceClass, error)
	ListServiceClasses(ctx context.Context, organizationID string) ([]v1.ServiceClass, error)
}

// NewServicesClient is used to create a new serviceInstances client
func NewServicesClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) *DefaultServicesClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultServicesClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultServicesClient defines the default services client
type DefaultServicesClient struct {
	baseClient

	client *swaggerclient.ServiceManager
	auth   runtime.ClientAuthInfoWriter
}

// CreateServiceInstance creates a service instance
func (c *DefaultServicesClient) CreateServiceInstance(ctx context.Context, organizationID string, instance *v1.ServiceInstance) (*v1.ServiceInstance, error) {
	params := serviceinstanceclient.AddServiceInstanceParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
		Body:         instance,
	}
	response, err := c.client.ServiceInstance.AddServiceInstance(&params, c.auth)
	if err != nil {
		return nil, createServiceInstanceSwaggerError(err)
	}
	return response.Payload, nil
}

func createServiceInstanceSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *serviceinstanceclient.AddServiceInstanceBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *serviceinstanceclient.AddServiceInstanceUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *serviceinstanceclient.AddServiceInstanceForbidden:
		return NewErrorForbidden(v.Payload)
	case *serviceinstanceclient.AddServiceInstanceConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *serviceinstanceclient.AddServiceInstanceDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteServiceInstance deletes a service instance
func (c *DefaultServicesClient) DeleteServiceInstance(ctx context.Context, organizationID string, serviceInstanceName string) error {
	params := serviceinstanceclient.DeleteServiceInstanceByNameParams{
		Context:             ctx,
		XDispatchOrg:        c.getOrgID(organizationID),
		ServiceInstanceName: serviceInstanceName,
	}
	_, err := c.client.ServiceInstance.DeleteServiceInstanceByName(&params, c.auth)
	if err != nil {
		return deleteServiceInstanceSwaggerError(err)
	}
	return nil
}

func deleteServiceInstanceSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *serviceinstanceclient.DeleteServiceInstanceByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *serviceinstanceclient.DeleteServiceInstanceByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *serviceinstanceclient.DeleteServiceInstanceByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *serviceinstanceclient.DeleteServiceInstanceByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *serviceinstanceclient.DeleteServiceInstanceByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetServiceInstance retrieves a service instance
func (c *DefaultServicesClient) GetServiceInstance(ctx context.Context, organizationID string, serviceInstanceName string) (*v1.ServiceInstance, error) {
	params := serviceinstanceclient.GetServiceInstanceByNameParams{
		Context:             ctx,
		XDispatchOrg:        c.getOrgID(organizationID),
		ServiceInstanceName: serviceInstanceName,
	}
	response, err := c.client.ServiceInstance.GetServiceInstanceByName(&params, c.auth)
	if err != nil {
		return nil, getServiceInstanceSwaggerError(err)
	}
	return response.Payload, nil
}

func getServiceInstanceSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *serviceinstanceclient.GetServiceInstanceByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *serviceinstanceclient.GetServiceInstanceByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *serviceinstanceclient.GetServiceInstanceByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *serviceinstanceclient.GetServiceInstanceByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *serviceinstanceclient.GetServiceInstanceByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListServiceInstances lists service instances
func (c *DefaultServicesClient) ListServiceInstances(ctx context.Context, organizationID string) ([]v1.ServiceInstance, error) {
	params := serviceinstanceclient.GetServiceInstancesParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.ServiceInstance.GetServiceInstances(&params, c.auth)
	if err != nil {
		return nil, listServiceInstancesSwaggerError(err)
	}
	var serviceInstances []v1.ServiceInstance
	for _, serviceInstance := range response.Payload {
		serviceInstances = append(serviceInstances, *serviceInstance)
	}
	return serviceInstances, nil
}

func listServiceInstancesSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *serviceinstanceclient.GetServiceInstancesUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *serviceinstanceclient.GetServiceInstancesForbidden:
		return NewErrorForbidden(v.Payload)
	case *serviceinstanceclient.GetServiceInstancesDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetServiceClass retrieves a service class
func (c *DefaultServicesClient) GetServiceClass(ctx context.Context, organizationID string, serviceClassName string) (*v1.ServiceClass, error) {
	params := serviceclassclient.GetServiceClassByNameParams{
		Context:          ctx,
		XDispatchOrg:     c.getOrgID(organizationID),
		ServiceClassName: serviceClassName,
	}
	response, err := c.client.ServiceClass.GetServiceClassByName(&params, c.auth)
	if err != nil {
		return nil, getServiceClassSwaggerError(err)
	}
	return response.Payload, nil
}

func getServiceClassSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *serviceclassclient.GetServiceClassByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *serviceclassclient.GetServiceClassByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *serviceclassclient.GetServiceClassByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *serviceclassclient.GetServiceClassByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *serviceclassclient.GetServiceClassByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListServiceClasses lists service classes
func (c *DefaultServicesClient) ListServiceClasses(ctx context.Context, organizationID string) ([]v1.ServiceClass, error) {
	params := serviceclassclient.GetServiceClassesParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.ServiceClass.GetServiceClasses(&params, c.auth)
	if err != nil {
		return nil, listServiceClassesSwaggerError(err)
	}
	serviceClasses := []v1.ServiceClass{}
	for _, serviceClass := range response.Payload {
		serviceClasses = append(serviceClasses, *serviceClass)
	}
	return serviceClasses, nil
}

func listServiceClassesSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *serviceclassclient.GetServiceClassesUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *serviceclassclient.GetServiceClassesForbidden:
		return NewErrorForbidden(v.Payload)
	case *serviceclassclient.GetServiceClassesDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
