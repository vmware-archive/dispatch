///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package drivers

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// NewAddDriverParams creates a new AddDriverParams object
// with the default values initialized.
func NewAddDriverParams() *AddDriverParams {
	var ()
	return &AddDriverParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewAddDriverParamsWithTimeout creates a new AddDriverParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewAddDriverParamsWithTimeout(timeout time.Duration) *AddDriverParams {
	var ()
	return &AddDriverParams{

		timeout: timeout,
	}
}

// NewAddDriverParamsWithContext creates a new AddDriverParams object
// with the default values initialized, and the ability to set a context for a request
func NewAddDriverParamsWithContext(ctx context.Context) *AddDriverParams {
	var ()
	return &AddDriverParams{

		Context: ctx,
	}
}

// NewAddDriverParamsWithHTTPClient creates a new AddDriverParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewAddDriverParamsWithHTTPClient(client *http.Client) *AddDriverParams {
	var ()
	return &AddDriverParams{
		HTTPClient: client,
	}
}

/*AddDriverParams contains all the parameters to send to the API endpoint
for the add driver operation typically these are written to a http.Request
*/
type AddDriverParams struct {

	/*Body
	  driver object

	*/
	Body *v1.EventDriver

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the add driver params
func (o *AddDriverParams) WithTimeout(timeout time.Duration) *AddDriverParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the add driver params
func (o *AddDriverParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the add driver params
func (o *AddDriverParams) WithContext(ctx context.Context) *AddDriverParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the add driver params
func (o *AddDriverParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the add driver params
func (o *AddDriverParams) WithHTTPClient(client *http.Client) *AddDriverParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the add driver params
func (o *AddDriverParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the add driver params
func (o *AddDriverParams) WithBody(body *v1.EventDriver) *AddDriverParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the add driver params
func (o *AddDriverParams) SetBody(body *v1.EventDriver) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *AddDriverParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
