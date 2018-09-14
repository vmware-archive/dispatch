///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations/endpoint"
)

// Handlers define a set of handlers for API Manager
type Handlers struct {
}

// APIHandlers is the interface for API gateway endpoints
type APIHandlers interface {
	AddAPI(params endpoint.AddAPIParams, principal interface{}) middleware.Responder
	DeleteAPI(params endpoint.DeleteAPIParams, principal interface{}) middleware.Responder
	UpdateAPI(params endpoint.UpdateAPIParams, principal interface{}) middleware.Responder
	GetAPI(params endpoint.GetAPIParams, principal interface{}) middleware.Responder
	GetAPIs(params endpoint.GetApisParams, principal interface{}) middleware.Responder
}

// NewHandlers create a new API Manager Handler
func NewHandlers() *Handlers {
	return &Handlers{}
}

// ConfigureHandlers configure handlers for API Manager
func (h *Handlers) ConfigureHandlers(routableAPI middleware.RoutableAPI, handlers APIHandlers) {
	a, ok := routableAPI.(*operations.APIManagerAPI)
	if !ok {
		panic("Cannot configure API-Manager API")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.Logger = log.Printf
	a.EndpointAddAPIHandler = endpoint.AddAPIHandlerFunc(handlers.AddAPI)
	a.EndpointDeleteAPIHandler = endpoint.DeleteAPIHandlerFunc(handlers.DeleteAPI)
	a.EndpointGetAPIHandler = endpoint.GetAPIHandlerFunc(handlers.GetAPI)
	a.EndpointGetApisHandler = endpoint.GetApisHandlerFunc(handlers.GetAPIs)
	a.EndpointUpdateAPIHandler = endpoint.UpdateAPIHandlerFunc(handlers.UpdateAPI)
}
