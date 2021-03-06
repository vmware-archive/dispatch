///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package image

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

// NewGetImageByNameParams creates a new GetImageByNameParams object
// with the default values initialized.
func NewGetImageByNameParams() GetImageByNameParams {

	var (
		// initialize parameters with default values

		xDispatchOrgDefault     = string("default")
		xDispatchProjectDefault = string("default")
	)

	return GetImageByNameParams{
		XDispatchOrg: &xDispatchOrgDefault,

		XDispatchProject: &xDispatchProjectDefault,
	}
}

// GetImageByNameParams contains all the bound params for the get image by name operation
// typically these are obtained from a http.Request
//
// swagger:parameters getImageByName
type GetImageByNameParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Pattern: ^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$
	  In: header
	  Default: "default"
	*/
	XDispatchOrg *string
	/*
	  Pattern: ^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$
	  In: header
	  Default: "default"
	*/
	XDispatchProject *string
	/*Name of image to return
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	ImageName string
	/*Filter on image tags
	  In: query
	  Collection Format: multi
	*/
	Tags []string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetImageByNameParams() beforehand.
func (o *GetImageByNameParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	if err := o.bindXDispatchOrg(r.Header[http.CanonicalHeaderKey("X-Dispatch-Org")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	if err := o.bindXDispatchProject(r.Header[http.CanonicalHeaderKey("X-Dispatch-Project")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	rImageName, rhkImageName, _ := route.Params.GetOK("imageName")
	if err := o.bindImageName(rImageName, rhkImageName, route.Formats); err != nil {
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

// bindXDispatchOrg binds and validates parameter XDispatchOrg from header.
func (o *GetImageByNameParams) bindXDispatchOrg(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetImageByNameParams()
		return nil
	}

	o.XDispatchOrg = &raw

	if err := o.validateXDispatchOrg(formats); err != nil {
		return err
	}

	return nil
}

// validateXDispatchOrg carries on validations for parameter XDispatchOrg
func (o *GetImageByNameParams) validateXDispatchOrg(formats strfmt.Registry) error {

	if err := validate.Pattern("X-Dispatch-Org", "header", (*o.XDispatchOrg), `^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$`); err != nil {
		return err
	}

	return nil
}

// bindXDispatchProject binds and validates parameter XDispatchProject from header.
func (o *GetImageByNameParams) bindXDispatchProject(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetImageByNameParams()
		return nil
	}

	o.XDispatchProject = &raw

	if err := o.validateXDispatchProject(formats); err != nil {
		return err
	}

	return nil
}

// validateXDispatchProject carries on validations for parameter XDispatchProject
func (o *GetImageByNameParams) validateXDispatchProject(formats strfmt.Registry) error {

	if err := validate.Pattern("X-Dispatch-Project", "header", (*o.XDispatchProject), `^[\w\d][\w\d\-]*[\w\d]|[\w\d]+$`); err != nil {
		return err
	}

	return nil
}

// bindImageName binds and validates parameter ImageName from path.
func (o *GetImageByNameParams) bindImageName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.ImageName = raw

	if err := o.validateImageName(formats); err != nil {
		return err
	}

	return nil
}

// validateImageName carries on validations for parameter ImageName
func (o *GetImageByNameParams) validateImageName(formats strfmt.Registry) error {

	if err := validate.Pattern("imageName", "path", o.ImageName, `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

// bindTags binds and validates array parameter Tags from query.
//
// Arrays are parsed according to CollectionFormat: "multi" (defaults to "csv" when empty).
func (o *GetImageByNameParams) bindTags(rawData []string, hasKey bool, formats strfmt.Registry) error {

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
