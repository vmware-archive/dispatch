///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/vmware/dispatch/pkg/images/backend"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	image "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/image"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type knHandlers struct {
	backend       backend.Backend
	httpClient    *http.Client
	namespace     string
	imageRegistry string
}

// NewHandlers is the constructor for image manager API knHandlers
func NewHandlers(kubecfgPath, namespace, imageRegistry string) Handlers {
	return &knHandlers{
		backend:       backend.KnativeBuild(kubecfgPath),
		httpClient:    &http.Client{},
		namespace:     namespace,
		imageRegistry: imageRegistry,
	}
}

func (h *knHandlers) addImage(params image.AddImageParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	img := params.Body
	utils.AdjustMeta(&img.Meta, dapi.Meta{Name: img.Name, Org: org, Project: project})

	log.Printf("adding name: %s, org:%s, proj:%s\n", img.Meta.Name, img.Meta.Org, img.Meta.Project)

	imageID := uuid.NewV4().String()
	img.ImageDestination = fmt.Sprintf("%s/%s", h.imageRegistry, imageID)

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

func (h *knHandlers) getImage(params image.GetImageByNameParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.ImageName
	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("getting image in %s:%s", org, project)
	img, err := h.backend.GetImage(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "get image"))
		return image.NewGetImageByNameDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String(err.Error()),
		})
	}

	return image.NewGetImageByNameOK().WithPayload(img)
}

func (h *knHandlers) deleteImage(params image.DeleteImageByNameParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.ImageName
	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("deleting image in %s:%s", org, project)
	err := h.backend.DeleteImage(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "delete image"))
		return image.NewDeleteImageByNameDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String(err.Error()),
		})
	}

	return image.NewDeleteImageByNameOK()
}

func (h *knHandlers) getImages(params image.GetImagesParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("getting images in %s:%s", org, project)
	dImages, err := h.backend.ListImage(ctx, &dapi.Meta{Org: org, Project: project})
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "listing images"))
		return image.NewGetImagesDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String(err.Error()),
		})
	}

	return image.NewGetImagesOK().WithPayload(dImages)
}

func (h *knHandlers) updateImage(params image.UpdateImageByNameParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	img := params.Body
	updated, err := h.backend.UpdateImage(ctx, img)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "updating image"))
		return image.NewUpdateImageByNameDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String(err.Error()),
		})
	}

	return image.NewUpdateImageByNameOK().WithPayload(updated)
}
