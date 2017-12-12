///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package subscriptions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetSubscriptionsHandlerFunc turns a function with the right signature into a get subscriptions handler
type GetSubscriptionsHandlerFunc func(GetSubscriptionsParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn GetSubscriptionsHandlerFunc) Handle(params GetSubscriptionsParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// GetSubscriptionsHandler interface for that can handle valid get subscriptions params
type GetSubscriptionsHandler interface {
	Handle(GetSubscriptionsParams, interface{}) middleware.Responder
}

// NewGetSubscriptions creates a new http.Handler for the get subscriptions operation
func NewGetSubscriptions(ctx *middleware.Context, handler GetSubscriptionsHandler) *GetSubscriptions {
	return &GetSubscriptions{Context: ctx, Handler: handler}
}

/*GetSubscriptions swagger:route GET /subscriptions subscriptions getSubscriptions

List all existing subscriptions

*/
type GetSubscriptions struct {
	Context *middleware.Context
	Handler GetSubscriptionsHandler
}

func (o *GetSubscriptions) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetSubscriptionsParams()

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
