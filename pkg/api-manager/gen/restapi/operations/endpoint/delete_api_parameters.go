///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package endpoint

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewDeleteAPIParams creates a new DeleteAPIParams object
// no default values defined in spec.
func NewDeleteAPIParams() DeleteAPIParams {

	return DeleteAPIParams{}
}

// DeleteAPIParams contains all the bound params for the delete API operation
// typically these are obtained from a http.Request
//
// swagger:parameters deleteAPI
type DeleteAPIParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: header
	*/
	XDISPATCHORGID string
	/*Name of API to work on
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	API string
	/*Filter based on tags
	  In: query
	  Collection Format: multi
	*/
	Tags []string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewDeleteAPIParams() beforehand.
func (o *DeleteAPIParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	if err := o.bindXDISPATCHORGID(r.Header[http.CanonicalHeaderKey("X-DISPATCH-ORG-ID")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	rAPI, rhkAPI, _ := route.Params.GetOK("api")
	if err := o.bindAPI(rAPI, rhkAPI, route.Formats); err != nil {
		res = append(res, err)
	}

	qTags, qhkTags, _ := qs.GetOK("tags")
	if err := o.bindTags(qTags, qhkTags, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *DeleteAPIParams) bindXDISPATCHORGID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("X-DISPATCH-ORG-ID", "header")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("X-DISPATCH-ORG-ID", "header", raw); err != nil {
		return err
	}

	o.XDISPATCHORGID = raw

	return nil
}

func (o *DeleteAPIParams) bindAPI(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

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

func (o *DeleteAPIParams) bindTags(rawData []string, hasKey bool, formats strfmt.Registry) error {

	// CollectionFormat: multi
	tagsIC := rawData

	if len(tagsIC) == 0 {
		return nil
	}

	var tagsIR []string
	for _, tagsIV := range tagsIC {
		tagsI := tagsIV

		tagsIR = append(tagsIR, tagsI)
	}

	o.Tags = tagsIR

	return nil
}
