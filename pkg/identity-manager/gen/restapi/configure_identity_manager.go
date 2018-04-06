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

	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/policy"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/serviceaccount"
)

//go:generate swagger generate server --target ../pkg/identity-manager/gen --name IdentityManager --spec ../swagger/identity-manager.yaml --exclude-main

func configureFlags(api *operations.IdentityManagerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.IdentityManagerAPI) http.Handler {
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

	api.PolicyAddPolicyHandler = policy.AddPolicyHandlerFunc(func(params policy.AddPolicyParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation policy.AddPolicy has not yet been implemented")
	})
	api.ServiceaccountAddServiceAccountHandler = serviceaccount.AddServiceAccountHandlerFunc(func(params serviceaccount.AddServiceAccountParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation serviceaccount.AddServiceAccount has not yet been implemented")
	})
	api.AuthHandler = operations.AuthHandlerFunc(func(params operations.AuthParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .Auth has not yet been implemented")
	})
	api.PolicyDeletePolicyHandler = policy.DeletePolicyHandlerFunc(func(params policy.DeletePolicyParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation policy.DeletePolicy has not yet been implemented")
	})
	api.ServiceaccountDeleteServiceAccountHandler = serviceaccount.DeleteServiceAccountHandlerFunc(func(params serviceaccount.DeleteServiceAccountParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation serviceaccount.DeleteServiceAccount has not yet been implemented")
	})
	api.PolicyGetPoliciesHandler = policy.GetPoliciesHandlerFunc(func(params policy.GetPoliciesParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation policy.GetPolicies has not yet been implemented")
	})
	api.PolicyGetPolicyHandler = policy.GetPolicyHandlerFunc(func(params policy.GetPolicyParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation policy.GetPolicy has not yet been implemented")
	})
	api.ServiceaccountGetServiceAccountHandler = serviceaccount.GetServiceAccountHandlerFunc(func(params serviceaccount.GetServiceAccountParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation serviceaccount.GetServiceAccount has not yet been implemented")
	})
	api.ServiceaccountGetServiceAccountsHandler = serviceaccount.GetServiceAccountsHandlerFunc(func(params serviceaccount.GetServiceAccountsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation serviceaccount.GetServiceAccounts has not yet been implemented")
	})
	api.HomeHandler = operations.HomeHandlerFunc(func(params operations.HomeParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .Home has not yet been implemented")
	})
	api.RedirectHandler = operations.RedirectHandlerFunc(func(params operations.RedirectParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .Redirect has not yet been implemented")
	})
	api.RootHandler = operations.RootHandlerFunc(func(params operations.RootParams) middleware.Responder {
		return middleware.NotImplemented("operation .Root has not yet been implemented")
	})
	api.PolicyUpdatePolicyHandler = policy.UpdatePolicyHandlerFunc(func(params policy.UpdatePolicyParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation policy.UpdatePolicy has not yet been implemented")
	})
	api.ServiceaccountUpdateServiceAccountHandler = serviceaccount.UpdateServiceAccountHandlerFunc(func(params serviceaccount.UpdateServiceAccountParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation serviceaccount.UpdateServiceAccount has not yet been implemented")
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
