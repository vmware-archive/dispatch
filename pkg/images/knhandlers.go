///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/vmware/dispatch/pkg/images/backend"
	baseimage "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/base_image"
	image "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/image"
)

type knHandlers struct {
	backend    backend.Backend
	httpClient *http.Client
	namespace  string
}

// NewHandlers is the constructor for image manager API knHandlers
func NewHandlers(kubecfgPath, namespace string) Handlers {
	return &knHandlers{
		backend:    backend.KnativeBuild(kubecfgPath),
		httpClient: &http.Client{},
		namespace:  namespace,
	}
}

// TODO: add base image handler
func (h *knHandlers) addBaseImage(params baseimage.AddBaseImageParams, principal interface{}) middleware.Responder {
	return nil
}

// TODO: add image handler
func (h *knHandlers) addImage(params image.AddImageParams, principal interface{}) middleware.Responder {
	return nil
}
