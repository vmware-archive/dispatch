///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package endpoint

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewDeleteAPIParams creates a new DeleteAPIParams object
// with the default values initialized.
func NewDeleteAPIParams() DeleteAPIParams {
	var ()
	return DeleteAPIParams{}
}

// DeleteAPIParams contains all the bound params for the delete API operation
// typically these are obtained from a http.Request
//
// swagger:parameters deleteAPI
type DeleteAPIParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*Name of API to work on
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	API string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *DeleteAPIParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	rAPI, rhkAPI, _ := route.Params.GetOK("api")
	if err := o.bindAPI(rAPI, rhkAPI, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *DeleteAPIParams) bindAPI(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	o.API = raw

	if err := o.validateAPI(formats); err != nil {
		return err
	}

	return nil
}

func (o *DeleteAPIParams) validateAPI(formats strfmt.Registry) error {

	if err := validate.Pattern("api", "path", o.API, `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}
