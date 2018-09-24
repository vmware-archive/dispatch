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
	baseimageclient "github.com/vmware/dispatch/pkg/images/gen/client/base_image"
	imageclient "github.com/vmware/dispatch/pkg/images/gen/client/image"
)

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
	response, err := c.client.Image.AddImage(&params)
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
	response, err := c.client.Image.DeleteImageByName(&params)
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
	response, err := c.client.Image.UpdateImageByName(&params)
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
	response, err := c.client.Image.GetImageByName(&params)
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
	response, err := c.client.Image.GetImages(&params)
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

// CreateBaseImage creates new base image
func (c *DefaultImagesClient) CreateBaseImage(ctx context.Context, organizationID string, image *v1.BaseImage) (*v1.BaseImage, error) {
	params := baseimageclient.AddBaseImageParams{
		Context:      ctx,
		Body:         image,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.AddBaseImage(&params)
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
func (c *DefaultImagesClient) DeleteBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error) {
	params := baseimageclient.DeleteBaseImageByNameParams{
		Context:       ctx,
		BaseImageName: baseImageName,
		XDispatchOrg:  swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.DeleteBaseImageByName(&params)
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
func (c *DefaultImagesClient) UpdateBaseImage(ctx context.Context, organizationID string, image *v1.BaseImage) (*v1.BaseImage, error) {
	params := baseimageclient.UpdateBaseImageByNameParams{
		Context:       ctx,
		Body:          image,
		BaseImageName: *image.Name,
		XDispatchOrg:  swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.UpdateBaseImageByName(&params)
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
func (c *DefaultImagesClient) GetBaseImage(ctx context.Context, organizationID string, baseImageName string) (*v1.BaseImage, error) {
	params := baseimageclient.GetBaseImageByNameParams{
		Context:       ctx,
		BaseImageName: baseImageName,
		XDispatchOrg:  swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.GetBaseImageByName(&params)
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
func (c *DefaultImagesClient) ListBaseImages(ctx context.Context, organizationID string) ([]v1.BaseImage, error) {
	params := baseimageclient.GetBaseImagesParams{
		Context:      ctx,
		XDispatchOrg: swag.String(c.getOrgID(organizationID)),
	}
	response, err := c.client.BaseImage.GetBaseImages(&params)
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
