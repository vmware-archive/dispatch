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
	swaggerclient "github.com/vmware/dispatch/pkg/image-manager/gen/client"
	baseimageclient "github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	imageclient "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
)

// NO TESTS

// ImagesClient defines the image client interface
type ImagesClient interface {
	// Images
	CreateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error)
	DeleteImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error)
	UpdateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error)
	GetImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error)
	ListImages(ctx context.Context, organizationID string) ([]v1.Image, error)

	// BaseImages
	CreateBaseImage(ctx context.Context, organizationID string, baseImage *v1.BaseImage) (*v1.BaseImage, error)
	DeleteBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error)
	UpdateBaseImage(ctx context.Context, organizationID string, baseImage *v1.BaseImage) (*v1.BaseImage, error)
	GetBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error)
	ListBaseImages(ctx context.Context, organizationID string) ([]v1.BaseImage, error)
}

// NewImagesClient is used to create a new Images client
func NewImagesClient(host string, auth runtime.ClientAuthInfoWriter) ImagesClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultImagesClient{
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// DefaultImagesClient defines the default images client
type DefaultImagesClient struct {
	client *swaggerclient.ImageManager
	auth   runtime.ClientAuthInfoWriter
}

// CreateImage creates new image
func (c *DefaultImagesClient) CreateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error) {
	params := imageclient.AddImageParams{
		Context:        ctx,
		Body:           image,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.Image.AddImage(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating the image")
	}
	return response.Payload, nil
}

// DeleteImage deletes an image
func (c *DefaultImagesClient) DeleteImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error) {
	params := imageclient.DeleteImageByNameParams{
		Context:        ctx,
		ImageName:      imageName,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.Image.DeleteImageByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when deleting the image")
	}
	return response.Payload, nil
}

// UpdateImage updates an image
func (c *DefaultImagesClient) UpdateImage(ctx context.Context, organizationID string, image *v1.Image) (*v1.Image, error) {
	params := imageclient.UpdateImageByNameParams{
		Context:        ctx,
		Body:           image,
		ImageName:      *image.Name,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.Image.UpdateImageByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when updating the image")
	}
	return response.Payload, nil
}

// GetImage retrieves an image
func (c *DefaultImagesClient) GetImage(ctx context.Context, organizationID string, imageName string) (*v1.Image, error) {
	params := imageclient.GetImageByNameParams{
		Context:        ctx,
		ImageName:      imageName,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.Image.GetImageByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when getting the image")
	}
	return response.Payload, nil
}

// ListImages returns a list of images
func (c *DefaultImagesClient) ListImages(ctx context.Context, organizationID string) ([]v1.Image, error) {
	params := imageclient.GetImagesParams{
		Context:        ctx,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.Image.GetImages(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when listing images")
	}
	var images []v1.Image
	for _, image := range response.Payload {
		images = append(images, *image)
	}
	return images, nil
}

// CreateBaseImage creates new base image
func (c *DefaultImagesClient) CreateBaseImage(ctx context.Context, organizationID string, image *v1.BaseImage) (*v1.BaseImage, error) {
	params := baseimageclient.AddBaseImageParams{
		Context:        ctx,
		Body:           image,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.BaseImage.AddBaseImage(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating the base image")
	}
	return response.Payload, nil
}

// DeleteBaseImage deletes the base image
func (c *DefaultImagesClient) DeleteBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error) {
	params := baseimageclient.DeleteBaseImageByNameParams{
		Context:        ctx,
		BaseImageName:  baseImageName,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.BaseImage.DeleteBaseImageByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when deleting the base image")
	}
	return response.Payload, nil
}

// UpdateBaseImage updates the base image
func (c *DefaultImagesClient) UpdateBaseImage(ctx context.Context, organizationID string, image *v1.BaseImage) (*v1.BaseImage, error) {
	params := baseimageclient.UpdateBaseImageByNameParams{
		Context:        ctx,
		Body:           image,
		BaseImageName:  *image.Name,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.BaseImage.UpdateBaseImageByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when updating the base image")
	}
	return response.Payload, nil
}

// GetBaseImage retrieves the base image
func (c *DefaultImagesClient) GetBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error) {
	params := baseimageclient.GetBaseImageByNameParams{
		Context:        ctx,
		BaseImageName:  baseImageName,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.BaseImage.GetBaseImageByName(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when retreiving the base image")
	}
	return response.Payload, nil
}

// ListBaseImages returns a list of base images
func (c *DefaultImagesClient) ListBaseImages(ctx context.Context, organizationID string) ([]v1.BaseImage, error) {
	params := baseimageclient.GetBaseImagesParams{
		Context:        ctx,
		XDISPATCHORGID: organizationID,
	}
	response, err := c.client.BaseImage.GetBaseImages(&params, c.auth)
	if err != nil {
		return nil, errors.Wrap(err, "error when listing base images")
	}
	var images []v1.BaseImage
	for _, image := range response.Payload {
		images = append(images, *image)
	}
	return images, nil
}
