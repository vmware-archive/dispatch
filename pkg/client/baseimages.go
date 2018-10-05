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
	swaggerclient "github.com/vmware/dispatch/pkg/baseimages/gen/client"
	baseimageclient "github.com/vmware/dispatch/pkg/baseimages/gen/client/base_image"
)

// BaseImagesClient defines the baseimage client interface
type BaseImagesClient interface {
	CreateBaseImage(ctx context.Context, organizationID string, baseImage *v1.BaseImage) (*v1.BaseImage, error)
	DeleteBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error)
	UpdateBaseImage(ctx context.Context, organizationID string, baseImage *v1.BaseImage) (*v1.BaseImage, error)
	GetBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error)
	ListBaseImages(ctx context.Context, organizationID string) ([]v1.BaseImage, error)
}

// NewBaseImagesClient is used to create a new BaseImages client
func NewBaseImagesClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) BaseImagesClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultBaseImagesClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultBaseImagesClient defines the default baseimages client
type DefaultBaseImagesClient struct {
	baseClient

	client *swaggerclient.BaseImages
	auth   runtime.ClientAuthInfoWriter
}

// CreateBaseImage creates new base image
func (c *DefaultBaseImagesClient) CreateBaseImage(ctx context.Context, organizationID string, image *v1.BaseImage) (*v1.BaseImage, error) {
	params := baseimageclient.AddBaseImageParams{
		Context:      ctx,
		Body:         image,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.AddBaseImage(&params, c.auth)
	if err != nil {
		return nil, createBaseImageSwaggerError(err)
	}
	return response.Payload, nil
}

func createBaseImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *baseimageclient.AddBaseImageBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *baseimageclient.AddBaseImageUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *baseimageclient.AddBaseImageForbidden:
		return NewErrorForbidden(v.Payload)
	case *baseimageclient.AddBaseImageConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *baseimageclient.AddBaseImageDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteBaseImage deletes the base image
func (c *DefaultBaseImagesClient) DeleteBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error) {
	params := baseimageclient.DeleteBaseImageByNameParams{
		Context:       ctx,
		BaseImageName: baseImageName,
		XDispatchOrg:  swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.DeleteBaseImageByName(&params, c.auth)
	if err != nil {
		return nil, deleteBaseImageSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteBaseImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *baseimageclient.DeleteBaseImageByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *baseimageclient.DeleteBaseImageByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *baseimageclient.DeleteBaseImageByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *baseimageclient.DeleteBaseImageByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *baseimageclient.DeleteBaseImageByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateBaseImage updates the base image
func (c *DefaultBaseImagesClient) UpdateBaseImage(ctx context.Context, organizationID string, image *v1.BaseImage) (*v1.BaseImage, error) {
	params := baseimageclient.UpdateBaseImageByNameParams{
		Context:       ctx,
		Body:          image,
		BaseImageName: image.Name,
		XDispatchOrg:  swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.UpdateBaseImageByName(&params, c.auth)
	if err != nil {
		return nil, updateBaseImageSwaggerError(err)
	}
	return response.Payload, nil
}

func updateBaseImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *baseimageclient.UpdateBaseImageByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *baseimageclient.UpdateBaseImageByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *baseimageclient.UpdateBaseImageByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *baseimageclient.UpdateBaseImageByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *baseimageclient.UpdateBaseImageByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetBaseImage retrieves the base image
func (c *DefaultBaseImagesClient) GetBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error) {
	params := baseimageclient.GetBaseImageByNameParams{
		Context:       ctx,
		BaseImageName: baseImageName,
		XDispatchOrg:  swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.GetBaseImageByName(&params, c.auth)
	if err != nil {
		return nil, getBaseImageSwaggerError(err)
	}
	return response.Payload, nil
}

func getBaseImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *baseimageclient.GetBaseImageByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *baseimageclient.GetBaseImageByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *baseimageclient.GetBaseImageByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *baseimageclient.GetBaseImageByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *baseimageclient.GetBaseImageByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListBaseImages returns a list of base images
func (c *DefaultBaseImagesClient) ListBaseImages(ctx context.Context, organizationID string) ([]v1.BaseImage, error) {
	params := baseimageclient.GetBaseImagesParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.GetBaseImages(&params, c.auth)
	if err != nil {
		return nil, listBaseImagesSwaggerError(err)
	}
	images := []v1.BaseImage{}
	for _, image := range response.Payload {
		images = append(images, *image)
	}
	return images, nil
}

func listBaseImagesSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *baseimageclient.GetBaseImagesUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *baseimageclient.GetBaseImagesForbidden:
		return NewErrorForbidden(v.Payload)
	case *baseimageclient.GetBaseImagesDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
