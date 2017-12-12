///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package image

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new image API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for image API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
AddImage adds a new image
*/
func (a *Client) AddImage(params *AddImageParams, authInfo runtime.ClientAuthInfoWriter) (*AddImageCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewAddImageParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "addImage",
		Method:             "POST",
		PathPattern:        "/",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &AddImageReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*AddImageCreated), nil

}

/*
DeleteImageByName deletes an image
*/
func (a *Client) DeleteImageByName(params *DeleteImageByNameParams, authInfo runtime.ClientAuthInfoWriter) (*DeleteImageByNameOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteImageByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteImageByName",
		Method:             "DELETE",
		PathPattern:        "/{imageName}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeleteImageByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*DeleteImageByNameOK), nil

}

/*
GetImageByName finds image by ID

Returns a single image
*/
func (a *Client) GetImageByName(params *GetImageByNameParams, authInfo runtime.ClientAuthInfoWriter) (*GetImageByNameOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetImageByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getImageByName",
		Method:             "GET",
		PathPattern:        "/{imageName}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetImageByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*GetImageByNameOK), nil

}

/*
GetImages gets all images

List all images
*/
func (a *Client) GetImages(params *GetImagesParams, authInfo runtime.ClientAuthInfoWriter) (*GetImagesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetImagesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getImages",
		Method:             "GET",
		PathPattern:        "/",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetImagesReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*GetImagesOK), nil

}

/*
UpdateImageByName updates an image
*/
func (a *Client) UpdateImageByName(params *UpdateImageByNameParams, authInfo runtime.ClientAuthInfoWriter) (*UpdateImageByNameOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateImageByNameParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "updateImageByName",
		Method:             "PUT",
		PathPattern:        "/{imageName}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UpdateImageByNameReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*UpdateImageByNameOK), nil

}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
