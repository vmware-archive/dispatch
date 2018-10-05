///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// AllInOneRouter implements a simple HTTP handler that routes requests to proper sub-service handlers
// When executing dispatch in a single binary mode.
type AllInOneRouter struct {
	EventsHandler     http.Handler
	FunctionsHandler  http.Handler
	SecretsHandler    http.Handler
	BaseImagesHandler http.Handler
	ImagesHandler     http.Handler
	IdentityHandler   http.Handler
	ServicesHandler   http.Handler
	EndpointsHandler  http.Handler
}

// ServeHTTP implements the http.Handler interface
func (d *AllInOneRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	components := strings.SplitN(path[1:], "/", 3)
	if len(components) < 2 {
		rw.Header().Add("Content-type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write(notFoundError())
		return
	}
	// version
	if components[0] != "v1" {
		rw.Header().Add("Content-type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write(notFoundError())
		return
	}
	switch components[1] {
	case "function", "runs":
		d.FunctionsHandler.ServeHTTP(rw, r)
	case "secret":
		d.SecretsHandler.ServeHTTP(rw, r)
	case "baseimage":
		d.BaseImagesHandler.ServeHTTP(rw, r)
	case "image":
		d.ImagesHandler.ServeHTTP(rw, r)
	case "event", "events":
		d.EventsHandler.ServeHTTP(rw, r)
	case "endpoint":
		d.EndpointsHandler.ServeHTTP(rw, r)
	case "iam", "eventdrivers", "application", "serviceclass", "serviceinstance":
		rw.Header().Add("Content-type", "application/json")
		rw.WriteHeader(http.StatusNotImplemented)
		rw.Write(notImplementedError())
	default:
		rw.Header().Add("Content-type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write(notFoundError())
	}
}

func notImplementedError() []byte {
	msg := "Endpoint is not implemented in this version of Dispatch"
	body := &v1.Error{
		Code:    501,
		Message: &msg,
	}
	response, _ := json.Marshal(body)
	return response
}

func notFoundError() []byte {
	msg := "resource does not exist"
	body := &v1.Error{
		Code:    404,
		Message: &msg,
	}
	response, _ := json.Marshal(body)
	return response
}
