///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package endpoint

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/api-manager/gen/models"
)

// NewUpdateAPIParams creates a new UpdateAPIParams object
// with the default values initialized.
func NewUpdateAPIParams() UpdateAPIParams {
	var ()
	return UpdateAPIParams{}
}

// UpdateAPIParams contains all the bound params for the update API operation
// typically these are obtained from a http.Request
//
// swagger:parameters updateAPI
type UpdateAPIParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*Name of API to work on
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	API string
	/*API object
	  Required: true
	  In: body
	*/
	Body *models.API
	/*Filter based on tags
	  In: query
	  Collection Format: multi
	*/
	Tags []string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *UpdateAPIParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	rAPI, rhkAPI, _ := route.Params.GetOK("api")
	if err := o.bindAPI(rAPI, rhkAPI, route.Formats); err != nil {
		res = append(res, err)
	}

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.API
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("body", "body"))
			} else {
				res = append(res, errors.NewParseError("body", "body", "", err))
			}

		} else {
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

	qTags, qhkTags, _ := qs.GetOK("tags")
	if err := o.bindTags(qTags, qhkTags, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *UpdateAPIParams) bindAPI(rawData []string, hasKey bool, formats strfmt.Registry) error {
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

func (o *UpdateAPIParams) validateAPI(formats strfmt.Registry) error {

	if err := validate.Pattern("api", "path", o.API, `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (o *UpdateAPIParams) bindTags(rawData []string, hasKey bool, formats strfmt.Registry) error {

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
