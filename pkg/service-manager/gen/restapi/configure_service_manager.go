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

	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations/service_class"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations/service_instance"
)

//go:generate swagger generate server --target ../pkg/service-manager/gen --name ServiceManager --spec ../swagger/service-manager.yaml --exclude-main

func configureFlags(api *operations.ServiceManagerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ServiceManagerAPI) http.Handler {
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

	api.ServiceInstanceAddServiceInstanceHandler = service_instance.AddServiceInstanceHandlerFunc(func(params service_instance.AddServiceInstanceParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation service_instance.AddServiceInstance has not yet been implemented")
	})
	api.ServiceInstanceDeleteServiceInstanceByNameHandler = service_instance.DeleteServiceInstanceByNameHandlerFunc(func(params service_instance.DeleteServiceInstanceByNameParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation service_instance.DeleteServiceInstanceByName has not yet been implemented")
	})
	api.ServiceClassGetServiceClassByNameHandler = service_class.GetServiceClassByNameHandlerFunc(func(params service_class.GetServiceClassByNameParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation service_class.GetServiceClassByName has not yet been implemented")
	})
	api.ServiceClassGetServiceClassesHandler = service_class.GetServiceClassesHandlerFunc(func(params service_class.GetServiceClassesParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation service_class.GetServiceClasses has not yet been implemented")
	})
	api.ServiceInstanceGetServiceInstanceByNameHandler = service_instance.GetServiceInstanceByNameHandlerFunc(func(params service_instance.GetServiceInstanceByNameParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation service_instance.GetServiceInstanceByName has not yet been implemented")
	})
	api.ServiceInstanceGetServiceInstancesHandler = service_instance.GetServiceInstancesHandlerFunc(func(params service_instance.GetServiceInstancesParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation service_instance.GetServiceInstances has not yet been implemented")
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
