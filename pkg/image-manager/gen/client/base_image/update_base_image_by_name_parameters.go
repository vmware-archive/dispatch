///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package base_image

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/swag"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

// NewUpdateBaseImageByNameParams creates a new UpdateBaseImageByNameParams object
// with the default values initialized.
func NewUpdateBaseImageByNameParams() *UpdateBaseImageByNameParams {
	var ()
	return &UpdateBaseImageByNameParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewUpdateBaseImageByNameParamsWithTimeout creates a new UpdateBaseImageByNameParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewUpdateBaseImageByNameParamsWithTimeout(timeout time.Duration) *UpdateBaseImageByNameParams {
	var ()
	return &UpdateBaseImageByNameParams{

		timeout: timeout,
	}
}

// NewUpdateBaseImageByNameParamsWithContext creates a new UpdateBaseImageByNameParams object
// with the default values initialized, and the ability to set a context for a request
func NewUpdateBaseImageByNameParamsWithContext(ctx context.Context) *UpdateBaseImageByNameParams {
	var ()
	return &UpdateBaseImageByNameParams{

		Context: ctx,
	}
}

// NewUpdateBaseImageByNameParamsWithHTTPClient creates a new UpdateBaseImageByNameParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewUpdateBaseImageByNameParamsWithHTTPClient(client *http.Client) *UpdateBaseImageByNameParams {
	var ()
	return &UpdateBaseImageByNameParams{
		HTTPClient: client,
	}
}

/*UpdateBaseImageByNameParams contains all the parameters to send to the API endpoint
for the update base image by name operation typically these are written to a http.Request
*/
type UpdateBaseImageByNameParams struct {

	/*BaseImageName
	  Name of base image to return

	*/
	BaseImageName string
	/*Body*/
	Body *models.BaseImage
	/*Tags
	  Filter based on tags

	*/
	Tags []string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the update base image by name params
func (o *UpdateBaseImageByNameParams) WithTimeout(timeout time.Duration) *UpdateBaseImageByNameParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the update base image by name params
func (o *UpdateBaseImageByNameParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the update base image by name params
func (o *UpdateBaseImageByNameParams) WithContext(ctx context.Context) *UpdateBaseImageByNameParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the update base image by name params
func (o *UpdateBaseImageByNameParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the update base image by name params
func (o *UpdateBaseImageByNameParams) WithHTTPClient(client *http.Client) *UpdateBaseImageByNameParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the update base image by name params
func (o *UpdateBaseImageByNameParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBaseImageName adds the baseImageName to the update base image by name params
func (o *UpdateBaseImageByNameParams) WithBaseImageName(baseImageName string) *UpdateBaseImageByNameParams {
	o.SetBaseImageName(baseImageName)
	return o
}

// SetBaseImageName adds the baseImageName to the update base image by name params
func (o *UpdateBaseImageByNameParams) SetBaseImageName(baseImageName string) {
	o.BaseImageName = baseImageName
}

// WithBody adds the body to the update base image by name params
func (o *UpdateBaseImageByNameParams) WithBody(body *models.BaseImage) *UpdateBaseImageByNameParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the update base image by name params
func (o *UpdateBaseImageByNameParams) SetBody(body *models.BaseImage) {
	o.Body = body
}

// WithTags adds the tags to the update base image by name params
func (o *UpdateBaseImageByNameParams) WithTags(tags []string) *UpdateBaseImageByNameParams {
	o.SetTags(tags)
	return o
}

// SetTags adds the tags to the update base image by name params
func (o *UpdateBaseImageByNameParams) SetTags(tags []string) {
	o.Tags = tags
}

// WriteToRequest writes these params to a swagger request
func (o *UpdateBaseImageByNameParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param baseImageName
	if err := r.SetPathParam("baseImageName", o.BaseImageName); err != nil {
		return err
	}

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	valuesTags := o.Tags

	joinedTags := swag.JoinByFormat(valuesTags, "multi")
	// query array param tags
	if err := r.SetQueryParam("tags", joinedTags...); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
