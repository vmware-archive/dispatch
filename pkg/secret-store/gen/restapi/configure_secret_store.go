///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	graceful "github.com/tylerb/graceful"

	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations/secret"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target ../root/go/src/github.com/vmware/dispatch/pkg/secret-store/gen --name SecretStore --spec ../root/go/src/github.com/vmware/dispatch/swagger/secret-store.yaml --exclude-main

func configureFlags(api *operations.SecretStoreAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.SecretStoreAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Cookie" header is set
	api.CookieAuth = func(token string) (interface{}, error) {
		return nil, errors.NotImplemented("api key auth (cookie) Cookie from header param [Cookie] has not yet been implemented")
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	api.SecretAddSecretHandler = secret.AddSecretHandlerFunc(func(params secret.AddSecretParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation secret.AddSecret has not yet been implemented")
	})
	api.SecretDeleteSecretHandler = secret.DeleteSecretHandlerFunc(func(params secret.DeleteSecretParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation secret.DeleteSecret has not yet been implemented")
	})
	api.SecretGetSecretHandler = secret.GetSecretHandlerFunc(func(params secret.GetSecretParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation secret.GetSecret has not yet been implemented")
	})
	api.SecretGetSecretsHandler = secret.GetSecretsHandlerFunc(func(params secret.GetSecretsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation secret.GetSecrets has not yet been implemented")
	})
	api.SecretUpdateSecretHandler = secret.UpdateSecretHandlerFunc(func(params secret.UpdateSecretParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation secret.UpdateSecret has not yet been implemented")
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
