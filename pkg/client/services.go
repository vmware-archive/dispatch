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

	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/service-manager/gen/client"
	serviceclassclient "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_class"
	serviceinstanceclient "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"
)

// ServicesClient defines the services client interface
type ServicesClient interface {
	// Service Instances
	CreateServiceInstance(ctx context.Context, serviceInstance *v1.ServiceInstance) (*v1.ServiceInstance, error)
	DeleteServiceInstance(ctx context.Context, serviceInstanceName string) error
	GetServiceInstance(ctx context.Context, serviceInstanceName string) (*v1.ServiceInstance, error)
	ListServiceInstances(ctx context.Context) ([]v1.ServiceInstance, error)

	// Service Classes
	GetServiceClass(ctx context.Context, serviceClassName string) (*v1.ServiceClass, error)
	ListServiceClasses(ctx context.Context) ([]v1.ServiceClass, error)
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
func (c *DefaultServicesClient) CreateServiceInstance(ctx context.Context, instance *v1.ServiceInstance) (*v1.ServiceInstance, error) {
	params := serviceinstanceclient.AddServiceInstanceParams{
		Context: ctx,
		Body:    instance,
	}
	response, err := c.client.ServiceInstance.AddServiceInstance(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating a service instance")
	}
	return response.Payload, nil
}

// DeleteServiceInstance deletes a service instance
func (c *DefaultServicesClient) DeleteServiceInstance(ctx context.Context, serviceInstanceName string) error {
	params := serviceinstanceclient.DeleteServiceInstanceByNameParams{
		Context:             ctx,
		ServiceInstanceName: serviceInstanceName,
	}
	_, err := c.client.ServiceInstance.DeleteServiceInstanceByName(&params, c.auth)
	if err != nil {
		return errors.Wrap(err, "error when deleting a service instance")
	}
	return nil
}

// GetServiceInstance retrieves a service instance
func (c *DefaultServicesClient) GetServiceInstance(ctx context.Context, serviceInstanceName string) (*v1.ServiceInstance, error) {
	params := serviceinstanceclient.GetServiceInstanceByNameParams{
		Context:             ctx,
		ServiceInstanceName: serviceInstanceName,
	}
	response, err := c.client.ServiceInstance.GetServiceInstanceByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retrieving a service instance")
	}
	return response.Payload, nil
}

// ListServiceInstances lists service instances
func (c *DefaultServicesClient) ListServiceInstances(ctx context.Context) ([]v1.ServiceInstance, error) {
	params := serviceinstanceclient.GetServiceInstancesParams{
		Context: ctx,
	}
	response, err := c.client.ServiceInstance.GetServiceInstances(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retrieving service instances")
	}
	var serviceInstances []v1.ServiceInstance
	for _, serviceInstance := range response.Payload {
		serviceInstances = append(serviceInstances, *serviceInstance)
	}
	return serviceInstances, nil
}

// GetServiceClass retrieves a service class
func (c *DefaultServicesClient) GetServiceClass(ctx context.Context, serviceClassName string) (*v1.ServiceClass, error) {
	params := serviceclassclient.GetServiceClassByNameParams{
		Context:          ctx,
		ServiceClassName: serviceClassName,
	}
	response, err := c.client.ServiceClass.GetServiceClassByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retrieving a service class")
	}
	return response.Payload, nil
}

// ListServiceClasses lists service classes
func (c *DefaultServicesClient) ListServiceClasses(ctx context.Context) ([]v1.ServiceClass, error) {
	params := serviceclassclient.GetServiceClassesParams{
		Context: ctx,
	}
	response, err := c.client.ServiceClass.GetServiceClasses(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retrieving a service class")
	}
	serviceClasses := []v1.ServiceClass{}
	for _, serviceClass := range response.Payload {
		serviceClasses = append(serviceClasses, *serviceClass)
	}
	return serviceClasses, nil
}
