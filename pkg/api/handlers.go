///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package api

// NO TESTS

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
)

// SwaggerAPI is an interface for APIs based on swagger documents
type SwaggerAPI interface {
	middleware.RoutableAPI
	SetSpec(*loads.Document)
}

// HandlerRegistrar is a function which is called to register handler functions
// to the API
type HandlerRegistrar func(middleware.RoutableAPI, entitystore.EntityStore)
