///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package base_image

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// DeleteBaseImageByNameHandlerFunc turns a function with the right signature into a delete base image by name handler
type DeleteBaseImageByNameHandlerFunc func(DeleteBaseImageByNameParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteBaseImageByNameHandlerFunc) Handle(params DeleteBaseImageByNameParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// DeleteBaseImageByNameHandler interface for that can handle valid delete base image by name params
type DeleteBaseImageByNameHandler interface {
	Handle(DeleteBaseImageByNameParams, interface{}) middleware.Responder
}

// NewDeleteBaseImageByName creates a new http.Handler for the delete base image by name operation
func NewDeleteBaseImageByName(ctx *middleware.Context, handler DeleteBaseImageByNameHandler) *DeleteBaseImageByName {
	return &DeleteBaseImageByName{Context: ctx, Handler: handler}
}

/*DeleteBaseImageByName swagger:route DELETE /base/{baseImageName} baseImage deleteBaseImageByName

Deletes a base image

*/
type DeleteBaseImageByName struct {
	Context *middleware.Context
	Handler DeleteBaseImageByNameHandler
}

func (o *DeleteBaseImageByName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDeleteBaseImageByNameParams()

	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		r = aCtx
	}
	var principal interface{}
	if uprinc != nil {
		principal = uprinc
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
