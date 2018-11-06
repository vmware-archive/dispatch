///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	graceful "github.com/tylerb/graceful"

	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/events"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
)

//go:generate swagger generate server --target ../pkg/event-manager/gen --name EventManager --spec ../swagger/event-manager.yaml --model-package v1 --skip-models --exclude-main

func configureFlags(api *operations.EventManagerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.EventManagerAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	api.BearerAuth = func(token string) (interface{}, error) {
		return nil, errors.NotImplemented("api key auth (bearer) Authorization from header param [Authorization] has not yet been implemented")
	}

	// Applies when the "Cookie" header is set
	api.CookieAuth = func(token string) (interface{}, error) {
		return nil, errors.NotImplemented("api key auth (cookie) Cookie from header param [Cookie] has not yet been implemented")
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	api.EventsIngestEventHandler = events.IngestEventHandlerFunc(func(params events.IngestEventParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation events.IngestEvent has not yet been implemented")
	})
	api.DriversAddDriverHandler = drivers.AddDriverHandlerFunc(func(params drivers.AddDriverParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.AddDriver has not yet been implemented")
	})
	api.DriversAddDriverTypeHandler = drivers.AddDriverTypeHandlerFunc(func(params drivers.AddDriverTypeParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.AddDriverType has not yet been implemented")
	})
	api.SubscriptionsAddSubscriptionHandler = subscriptions.AddSubscriptionHandlerFunc(func(params subscriptions.AddSubscriptionParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation subscriptions.AddSubscription has not yet been implemented")
	})
	api.DriversDeleteDriverHandler = drivers.DeleteDriverHandlerFunc(func(params drivers.DeleteDriverParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.DeleteDriver has not yet been implemented")
	})
	api.DriversDeleteDriverTypeHandler = drivers.DeleteDriverTypeHandlerFunc(func(params drivers.DeleteDriverTypeParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.DeleteDriverType has not yet been implemented")
	})
	api.SubscriptionsDeleteSubscriptionHandler = subscriptions.DeleteSubscriptionHandlerFunc(func(params subscriptions.DeleteSubscriptionParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation subscriptions.DeleteSubscription has not yet been implemented")
	})
	api.EventsEmitEventHandler = events.EmitEventHandlerFunc(func(params events.EmitEventParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation events.EmitEvent has not yet been implemented")
	})
	api.DriversGetDriverHandler = drivers.GetDriverHandlerFunc(func(params drivers.GetDriverParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.GetDriver has not yet been implemented")
	})
	api.DriversGetDriverTypeHandler = drivers.GetDriverTypeHandlerFunc(func(params drivers.GetDriverTypeParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.GetDriverType has not yet been implemented")
	})
	api.DriversGetDriverTypesHandler = drivers.GetDriverTypesHandlerFunc(func(params drivers.GetDriverTypesParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.GetDriverTypes has not yet been implemented")
	})
	api.DriversGetDriversHandler = drivers.GetDriversHandlerFunc(func(params drivers.GetDriversParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.GetDrivers has not yet been implemented")
	})
	api.SubscriptionsGetSubscriptionHandler = subscriptions.GetSubscriptionHandlerFunc(func(params subscriptions.GetSubscriptionParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation subscriptions.GetSubscription has not yet been implemented")
	})
	api.SubscriptionsGetSubscriptionsHandler = subscriptions.GetSubscriptionsHandlerFunc(func(params subscriptions.GetSubscriptionsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation subscriptions.GetSubscriptions has not yet been implemented")
	})
	api.DriversUpdateDriverHandler = drivers.UpdateDriverHandlerFunc(func(params drivers.UpdateDriverParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.UpdateDriver has not yet been implemented")
	})
	api.DriversUpdateDriverTypeHandler = drivers.UpdateDriverTypeHandlerFunc(func(params drivers.UpdateDriverTypeParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation drivers.UpdateDriverType has not yet been implemented")
	})
	api.SubscriptionsUpdateSubscriptionHandler = subscriptions.UpdateSubscriptionHandlerFunc(func(params subscriptions.UpdateSubscriptionParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation subscriptions.UpdateSubscription has not yet been implemented")
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
