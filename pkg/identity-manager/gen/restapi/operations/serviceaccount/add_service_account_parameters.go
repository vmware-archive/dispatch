///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package serviceaccount

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// NewAddServiceAccountParams creates a new AddServiceAccountParams object
// no default values defined in spec.
func NewAddServiceAccountParams() AddServiceAccountParams {

	return AddServiceAccountParams{}
}

// AddServiceAccountParams contains all the bound params for the add service account operation
// typically these are obtained from a http.Request
//
// swagger:parameters addServiceAccount
type AddServiceAccountParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Service Account Object
	  Required: true
	  In: body
	*/
	Body *v1.ServiceAccount
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewAddServiceAccountParams() beforehand.
func (o *AddServiceAccountParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body v1.ServiceAccount
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("body", "body"))
			} else {
				res = append(res, errors.NewParseError("body", "body", "", err))
			}
		} else {

			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Body = &body
			}
		}
	} else {
		res = append(res, errors.Required("body", "body"))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
