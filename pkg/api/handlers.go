///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package api

// NO TESTS

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
)

// SwaggerAPI is an interface for APIs based on swagger documents
type SwaggerAPI interface {
	middleware.RoutableAPI
	SetSpec(*loads.Document)
}

// HandlerRegistrar is a function which is called to register handler functions
// to the API
type HandlerRegistrar func(middleware.RoutableAPI)
