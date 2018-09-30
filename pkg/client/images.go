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
	swaggerclient "github.com/vmware/dispatch/pkg/images/gen/client"
	imageclient "github.com/vmware/dispatch/pkg/images/gen/client/image"
)

// ImagesClient defines the image client interface
type ImagesClient interface {
	CreateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error)
	DeleteImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error)
	UpdateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error)
	GetImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error)
	ListImages(ctx context.Context, organizationID string) ([]v1.Image, error)
}

// NewImagesClient is used to create a new Images client
func NewImagesClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) ImagesClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultImagesClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultImagesClient defines the default images client
type DefaultImagesClient struct {
	baseClient

	client *swaggerclient.Images
	auth   runtime.ClientAuthInfoWriter
}

// CreateImage creates new image
func (c *DefaultImagesClient) CreateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error) {
	params := imageclient.AddImageParams{
		Context:      ctx,
		Body:         image,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Image.AddImage(&params, c.auth)
	if err != nil {
		return nil, createImageSwaggerError(err)
	}
	return response.Payload, nil
}

func createImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *imageclient.AddImageBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *imageclient.AddImageUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *imageclient.AddImageForbidden:
		return NewErrorForbidden(v.Payload)
	case *imageclient.AddImageConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *imageclient.AddImageDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteImage deletes an image
func (c *DefaultImagesClient) DeleteImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error) {
	params := imageclient.DeleteImageByNameParams{
		Context:      ctx,
		ImageName:    imageName,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Image.DeleteImageByName(&params, c.auth)
	if err != nil {
		return nil, deleteImageSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *imageclient.DeleteImageByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *imageclient.DeleteImageByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *imageclient.DeleteImageByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *imageclient.DeleteImageByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *imageclient.DeleteImageByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateImage updates an image
func (c *DefaultImagesClient) UpdateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error) {
	params := imageclient.UpdateImageByNameParams{
		Context:      ctx,
		Body:         image,
		ImageName:    image.Name,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Image.UpdateImageByName(&params, c.auth)
	if err != nil {
		return nil, updateImageSwaggerError(err)
	}
	return response.Payload, nil
}

func updateImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *imageclient.UpdateImageByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *imageclient.UpdateImageByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *imageclient.UpdateImageByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *imageclient.UpdateImageByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *imageclient.UpdateImageByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetImage retrieves an image
func (c *DefaultImagesClient) GetImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error) {
	params := imageclient.GetImageByNameParams{
		Context:      ctx,
		ImageName:    imageName,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Image.GetImageByName(&params, c.auth)
	if err != nil {
		return nil, getImageSwaggerError(err)
	}
	return response.Payload, nil
}

func getImageSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *imageclient.GetImageByNameBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *imageclient.GetImageByNameUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *imageclient.GetImageByNameForbidden:
		return NewErrorForbidden(v.Payload)
	case *imageclient.GetImageByNameNotFound:
		return NewErrorNotFound(v.Payload)
	case *imageclient.GetImageByNameDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListImages returns a list of images
func (c *DefaultImagesClient) ListImages(ctx context.Context, organizationID string) ([]v1.Image, error) {
	params := imageclient.GetImagesParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.Image.GetImages(&params, c.auth)
	if err != nil {
		return nil, listImagesSwaggerError(err)
	}
	images := []v1.Image{}
	for _, image := range response.Payload {
		images = append(images, *image)
	}
	return images, nil
}

func listImagesSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *imageclient.GetImagesBadRequest:
		return NewErrorServerUnknownError(v.Payload)
	case *imageclient.GetImagesUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *imageclient.GetImagesForbidden:
		return NewErrorForbidden(v.Payload)
	case *imageclient.GetImagesDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
