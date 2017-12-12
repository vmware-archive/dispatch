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

// UpdateBaseImageByNameHandlerFunc turns a function with the right signature into a update base image by name handler
type UpdateBaseImageByNameHandlerFunc func(UpdateBaseImageByNameParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn UpdateBaseImageByNameHandlerFunc) Handle(params UpdateBaseImageByNameParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// UpdateBaseImageByNameHandler interface for that can handle valid update base image by name params
type UpdateBaseImageByNameHandler interface {
	Handle(UpdateBaseImageByNameParams, interface{}) middleware.Responder
}

// NewUpdateBaseImageByName creates a new http.Handler for the update base image by name operation
func NewUpdateBaseImageByName(ctx *middleware.Context, handler UpdateBaseImageByNameHandler) *UpdateBaseImageByName {
	return &UpdateBaseImageByName{Context: ctx, Handler: handler}
}

/*UpdateBaseImageByName swagger:route PUT /base/{baseImageName} baseImage updateBaseImageByName

Updates a base image

*/
type UpdateBaseImageByName struct {
	Context *middleware.Context
	Handler UpdateBaseImageByNameHandler
}

func (o *UpdateBaseImageByName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewUpdateBaseImageByNameParams()

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
