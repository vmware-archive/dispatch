///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package image

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetImagesHandlerFunc turns a function with the right signature into a get images handler
type GetImagesHandlerFunc func(GetImagesParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn GetImagesHandlerFunc) Handle(params GetImagesParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// GetImagesHandler interface for that can handle valid get images params
type GetImagesHandler interface {
	Handle(GetImagesParams, interface{}) middleware.Responder
}

// NewGetImages creates a new http.Handler for the get images operation
func NewGetImages(ctx *middleware.Context, handler GetImagesHandler) *GetImages {
	return &GetImages{Context: ctx, Handler: handler}
}

/*GetImages swagger:route GET / image getImages

Get all images

List all images

*/
type GetImages struct {
	Context *middleware.Context
	Handler GetImagesHandler
}

func (o *GetImages) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetImagesParams()

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
