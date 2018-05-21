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

// NewGetImagesParams creates a new GetImagesParams object
// no default values defined in spec.
func NewGetImagesParams() GetImagesParams {

	return GetImagesParams{}
}

// GetImagesParams contains all the bound params for the get images operation
// typically these are obtained from a http.Request
//
// swagger:parameters getImages
type GetImagesParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: header
	*/
	XDISPATCHORGID string
	/*image runtime language
	  In: query
	*/
	Language *string
	/*Filter on image tags
	  In: query
	  Collection Format: multi
	*/
	Tags []string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetImagesParams() beforehand.
func (o *GetImagesParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	if err := o.bindXDISPATCHORGID(r.Header[http.CanonicalHeaderKey("X-DISPATCH-ORG-ID")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	qLanguage, qhkLanguage, _ := qs.GetOK("language")
	if err := o.bindLanguage(qLanguage, qhkLanguage, route.Formats); err != nil {
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

func (o *GetImagesParams) bindXDISPATCHORGID(rawData []string, hasKey bool, formats strfmt.Registry) error {
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

func (o *GetImagesParams) bindLanguage(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false
	if raw == "" { // empty values pass all other validations
		return nil
	}

	o.Language = &raw

	return nil
}

func (o *GetImagesParams) bindTags(rawData []string, hasKey bool, formats strfmt.Registry) error {

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
