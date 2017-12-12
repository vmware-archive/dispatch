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

// GetBaseImageByNameHandlerFunc turns a function with the right signature into a get base image by name handler
type GetBaseImageByNameHandlerFunc func(GetBaseImageByNameParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn GetBaseImageByNameHandlerFunc) Handle(params GetBaseImageByNameParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// GetBaseImageByNameHandler interface for that can handle valid get base image by name params
type GetBaseImageByNameHandler interface {
	Handle(GetBaseImageByNameParams, interface{}) middleware.Responder
}

// NewGetBaseImageByName creates a new http.Handler for the get base image by name operation
func NewGetBaseImageByName(ctx *middleware.Context, handler GetBaseImageByNameHandler) *GetBaseImageByName {
	return &GetBaseImageByName{Context: ctx, Handler: handler}
}

/*GetBaseImageByName swagger:route GET /base/{baseImageName} baseImage getBaseImageByName

Find base image by Name

Returns a single base image

*/
type GetBaseImageByName struct {
	Context *middleware.Context
	Handler GetBaseImageByNameHandler
}

func (o *GetBaseImageByName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetBaseImageByNameParams()

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
