///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"
	"strings"

	errors "github.com/go-openapi/errors"
	loads "github.com/go-openapi/loads"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	security "github.com/go-openapi/runtime/security"
	spec "github.com/go-openapi/spec"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/baseimages/gen/restapi/operations/base_image"
)

// NewBaseImagesAPI creates a new BaseImages instance
func NewBaseImagesAPI(spec *loads.Document) *BaseImagesAPI {
	return &BaseImagesAPI{
		handlers:            make(map[string]map[string]http.Handler),
		formats:             strfmt.Default,
		defaultConsumes:     "application/json",
		defaultProduces:     "application/json",
		customConsumers:     make(map[string]runtime.Consumer),
		customProducers:     make(map[string]runtime.Producer),
		ServerShutdown:      func() {},
		spec:                spec,
		ServeError:          errors.ServeError,
		BasicAuthenticator:  security.BasicAuth,
		APIKeyAuthenticator: security.APIKeyAuth,
		BearerAuthenticator: security.BearerAuth,
		JSONConsumer:        runtime.JSONConsumer(),
		JSONProducer:        runtime.JSONProducer(),
		BaseImageAddBaseImageHandler: base_image.AddBaseImageHandlerFunc(func(params base_image.AddBaseImageParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation BaseImageAddBaseImage has not yet been implemented")
		}),
		BaseImageDeleteBaseImageByNameHandler: base_image.DeleteBaseImageByNameHandlerFunc(func(params base_image.DeleteBaseImageByNameParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation BaseImageDeleteBaseImageByName has not yet been implemented")
		}),
		BaseImageGetBaseImageByNameHandler: base_image.GetBaseImageByNameHandlerFunc(func(params base_image.GetBaseImageByNameParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation BaseImageGetBaseImageByName has not yet been implemented")
		}),
		BaseImageGetBaseImagesHandler: base_image.GetBaseImagesHandlerFunc(func(params base_image.GetBaseImagesParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation BaseImageGetBaseImages has not yet been implemented")
		}),
		BaseImageUpdateBaseImageByNameHandler: base_image.UpdateBaseImageByNameHandlerFunc(func(params base_image.UpdateBaseImageByNameParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation BaseImageUpdateBaseImageByName has not yet been implemented")
		}),

		// Applies when the "Authorization" header is set
		BearerAuth: func(token string) (interface{}, error) {
			return nil, errors.NotImplemented("api key auth (bearer) Authorization from header param [Authorization] has not yet been implemented")
		},
		// Applies when the "Cookie" header is set
		CookieAuth: func(token string) (interface{}, error) {
			return nil, errors.NotImplemented("api key auth (cookie) Cookie from header param [Cookie] has not yet been implemented")
		},

		// default authorizer is authorized meaning no requests are blocked
		APIAuthorizer: security.Authorized(),
	}
}

/*BaseImagesAPI VMware Dispatch BaseImages
 */
type BaseImagesAPI struct {
	spec            *loads.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	customConsumers map[string]runtime.Consumer
	customProducers map[string]runtime.Producer
	defaultConsumes string
	defaultProduces string
	Middleware      func(middleware.Builder) http.Handler

	// BasicAuthenticator generates a runtime.Authenticator from the supplied basic auth function.
	// It has a default implemention in the security package, however you can replace it for your particular usage.
	BasicAuthenticator func(security.UserPassAuthentication) runtime.Authenticator
	// APIKeyAuthenticator generates a runtime.Authenticator from the supplied token auth function.
	// It has a default implemention in the security package, however you can replace it for your particular usage.
	APIKeyAuthenticator func(string, string, security.TokenAuthentication) runtime.Authenticator
	// BearerAuthenticator generates a runtime.Authenticator from the supplied bearer token auth function.
	// It has a default implemention in the security package, however you can replace it for your particular usage.
	BearerAuthenticator func(string, security.ScopedTokenAuthentication) runtime.Authenticator

	// JSONConsumer registers a consumer for a "application/json" mime type
	JSONConsumer runtime.Consumer

	// JSONProducer registers a producer for a "application/json" mime type
	JSONProducer runtime.Producer

	// BearerAuth registers a function that takes a token and returns a principal
	// it performs authentication based on an api key Authorization provided in the header
	BearerAuth func(string) (interface{}, error)

	// CookieAuth registers a function that takes a token and returns a principal
	// it performs authentication based on an api key Cookie provided in the header
	CookieAuth func(string) (interface{}, error)

	// APIAuthorizer provides access control (ACL/RBAC/ABAC) by providing access to the request and authenticated principal
	APIAuthorizer runtime.Authorizer

	// BaseImageAddBaseImageHandler sets the operation handler for the add base image operation
	BaseImageAddBaseImageHandler base_image.AddBaseImageHandler
	// BaseImageDeleteBaseImageByNameHandler sets the operation handler for the delete base image by name operation
	BaseImageDeleteBaseImageByNameHandler base_image.DeleteBaseImageByNameHandler
	// BaseImageGetBaseImageByNameHandler sets the operation handler for the get base image by name operation
	BaseImageGetBaseImageByNameHandler base_image.GetBaseImageByNameHandler
	// BaseImageGetBaseImagesHandler sets the operation handler for the get base images operation
	BaseImageGetBaseImagesHandler base_image.GetBaseImagesHandler
	// BaseImageUpdateBaseImageByNameHandler sets the operation handler for the update base image by name operation
	BaseImageUpdateBaseImageByNameHandler base_image.UpdateBaseImageByNameHandler

	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup

	// User defined logger function.
	Logger func(string, ...interface{})
}

// SetDefaultProduces sets the default produces media type
func (o *BaseImagesAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *BaseImagesAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// SetSpec sets a spec that will be served for the clients.
func (o *BaseImagesAPI) SetSpec(spec *loads.Document) {
	o.spec = spec
}

// DefaultProduces returns the default produces media type
func (o *BaseImagesAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *BaseImagesAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *BaseImagesAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *BaseImagesAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the BaseImagesAPI
func (o *BaseImagesAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}

	if o.BearerAuth == nil {
		unregistered = append(unregistered, "AuthorizationAuth")
	}

	if o.CookieAuth == nil {
		unregistered = append(unregistered, "CookieAuth")
	}

	if o.BaseImageAddBaseImageHandler == nil {
		unregistered = append(unregistered, "base_image.AddBaseImageHandler")
	}

	if o.BaseImageDeleteBaseImageByNameHandler == nil {
		unregistered = append(unregistered, "base_image.DeleteBaseImageByNameHandler")
	}

	if o.BaseImageGetBaseImageByNameHandler == nil {
		unregistered = append(unregistered, "base_image.GetBaseImageByNameHandler")
	}

	if o.BaseImageGetBaseImagesHandler == nil {
		unregistered = append(unregistered, "base_image.GetBaseImagesHandler")
	}

	if o.BaseImageUpdateBaseImageByNameHandler == nil {
		unregistered = append(unregistered, "base_image.UpdateBaseImageByNameHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *BaseImagesAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *BaseImagesAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]runtime.Authenticator {

	result := make(map[string]runtime.Authenticator)
	for name, scheme := range schemes {
		switch name {

		case "bearer":

			result[name] = o.APIKeyAuthenticator(scheme.Name, scheme.In, o.BearerAuth)

		case "cookie":

			result[name] = o.APIKeyAuthenticator(scheme.Name, scheme.In, o.CookieAuth)

		}
	}
	return result

}

// Authorizer returns the registered authorizer
func (o *BaseImagesAPI) Authorizer() runtime.Authorizer {

	return o.APIAuthorizer

}

// ConsumersFor gets the consumers for the specified media types
func (o *BaseImagesAPI) ConsumersFor(mediaTypes []string) map[string]runtime.Consumer {

	result := make(map[string]runtime.Consumer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONConsumer

		}

		if c, ok := o.customConsumers[mt]; ok {
			result[mt] = c
		}
	}
	return result

}

// ProducersFor gets the producers for the specified media types
func (o *BaseImagesAPI) ProducersFor(mediaTypes []string) map[string]runtime.Producer {

	result := make(map[string]runtime.Producer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONProducer

		}

		if p, ok := o.customProducers[mt]; ok {
			result[mt] = p
		}
	}
	return result

}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *BaseImagesAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	if path == "/" {
		path = ""
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

// Context returns the middleware context for the base images API
func (o *BaseImagesAPI) Context() *middleware.Context {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	return o.context
}

func (o *BaseImagesAPI) initHandlerCache() {
	o.Context() // don't care about the result, just that the initialization happened

	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"][""] = base_image.NewAddBaseImage(o.context, o.BaseImageAddBaseImageHandler)

	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/{baseImageName}"] = base_image.NewDeleteBaseImageByName(o.context, o.BaseImageDeleteBaseImageByNameHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/{baseImageName}"] = base_image.NewGetBaseImageByName(o.context, o.BaseImageGetBaseImageByNameHandler)

	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"][""] = base_image.NewGetBaseImages(o.context, o.BaseImageGetBaseImagesHandler)

	if o.handlers["PUT"] == nil {
		o.handlers["PUT"] = make(map[string]http.Handler)
	}
	o.handlers["PUT"]["/{baseImageName}"] = base_image.NewUpdateBaseImageByName(o.context, o.BaseImageUpdateBaseImageByNameHandler)

}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *BaseImagesAPI) Serve(builder middleware.Builder) http.Handler {
	o.Init()

	if o.Middleware != nil {
		return o.Middleware(builder)
	}
	return o.context.APIHandler(builder)
}

// Init allows you to just initialize the handler cache, you can then recompose the middleware as you see fit
func (o *BaseImagesAPI) Init() {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}
}

// RegisterConsumer allows you to add (or override) a consumer for a media type.
func (o *BaseImagesAPI) RegisterConsumer(mediaType string, consumer runtime.Consumer) {
	o.customConsumers[mediaType] = consumer
}

// RegisterProducer allows you to add (or override) a producer for a media type.
func (o *BaseImagesAPI) RegisterProducer(mediaType string, producer runtime.Producer) {
	o.customProducers[mediaType] = producer
}