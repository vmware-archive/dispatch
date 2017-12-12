///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package events

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// EmitEventHandlerFunc turns a function with the right signature into a emit event handler
type EmitEventHandlerFunc func(EmitEventParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn EmitEventHandlerFunc) Handle(params EmitEventParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// EmitEventHandler interface for that can handle valid emit event params
type EmitEventHandler interface {
	Handle(EmitEventParams, interface{}) middleware.Responder
}

// NewEmitEvent creates a new http.Handler for the emit event operation
func NewEmitEvent(ctx *middleware.Context, handler EmitEventHandler) *EmitEvent {
	return &EmitEvent{Context: ctx, Handler: handler}
}

/*EmitEvent swagger:route POST / events emitEvent

Emit an event

*/
type EmitEvent struct {
	Context *middleware.Context
	Handler EmitEventHandler
}

func (o *EmitEvent) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewEmitEventParams()

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
