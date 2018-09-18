///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	"github.com/vmware/dispatch/pkg/images/backend"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	baseimage "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/base_image"
	image "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/image"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"

	log "github.com/sirupsen/logrus"
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
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	baseimg := params.Body
	utils.AdjustMeta(&baseimg.Meta, dapi.Meta{Org: org, Project: project})

	createdBaseimg, err := h.backend.AddBaseImage(ctx, baseimg)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "creating a base image"))
		return baseimage.NewAddBaseImageDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("base-image", baseimg.Meta.Name),
		})
	}
	return baseimage.NewAddBaseImageCreated().WithPayload(createdBaseimg)
}

// TODO: add image handler
func (h *knHandlers) addImage(params image.AddImageParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	img := params.Body
	utils.AdjustMeta(&img.Meta, dapi.Meta{Org: org, Project: project})

	createdImage, err := h.backend.AddImage(ctx, img)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "creating a image"))
		return image.NewAddImageDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("image", img.Meta.Name),
		})
	}
	return image.NewAddImageCreated().WithPayload(createdImage)
}
