///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package application

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewDeleteAppParams creates a new DeleteAppParams object
// no default values defined in spec.
func NewDeleteAppParams() DeleteAppParams {

	return DeleteAppParams{}
}

// DeleteAppParams contains all the bound params for the delete app operation
// typically these are obtained from a http.Request
//
// swagger:parameters deleteApp
type DeleteAppParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: header
	*/
	XDispatchOrg string
	/*Name of Application to work on
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	Application string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewDeleteAppParams() beforehand.
func (o *DeleteAppParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := o.bindXDispatchOrg(r.Header[http.CanonicalHeaderKey("X-Dispatch-Org")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	rApplication, rhkApplication, _ := route.Params.GetOK("application")
	if err := o.bindApplication(rApplication, rhkApplication, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *DeleteAppParams) bindXDispatchOrg(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("X-Dispatch-Org", "header")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("X-Dispatch-Org", "header", raw); err != nil {
		return err
	}

	o.XDispatchOrg = raw

	return nil
}

func (o *DeleteAppParams) bindApplication(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.Application = raw

	if err := o.validateApplication(formats); err != nil {
		return err
	}

	return nil
}

func (o *DeleteAppParams) validateApplication(formats strfmt.Registry) error {

	if err := validate.Pattern("application", "path", o.Application, `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}
