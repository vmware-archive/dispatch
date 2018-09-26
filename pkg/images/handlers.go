///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	log "github.com/sirupsen/logrus"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	derrors "github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/images/backend"
	"github.com/vmware/dispatch/pkg/images/gen/restapi/operations"
	image "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/image"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Handlers interface declares methods for image-manage API
// pricinpal interface{} reserved for security authentication
type Handlers interface {
	addImage(params image.AddImageParams) middleware.Responder
	getImage(params image.GetImageByNameParams) middleware.Responder
	deleteImage(params image.DeleteImageByNameParams) middleware.Responder
	getImages(params image.GetImagesParams) middleware.Responder
	updateImage(params image.UpdateImageByNameParams) middleware.Responder
}

// ConfigureHandlers registers the image manager handlers to API
func ConfigureHandlers(api middleware.RoutableAPI, h Handlers) {
	a, ok := api.(*operations.ImagesAPI)
	if !ok {
		panic("Cannot configure image manager apis")
	}

	// TODO: authentication CookieAuth/BearerAuth

	a.Logger = log.Printf

	a.ImageAddImageHandler = image.AddImageHandlerFunc(h.addImage)
	a.ImageGetImageByNameHandler = image.GetImageByNameHandlerFunc(h.getImage)
	a.ImageDeleteImageByNameHandler = image.DeleteImageByNameHandlerFunc(h.deleteImage)
	a.ImageGetImagesHandler = image.GetImagesHandlerFunc(h.getImages)
	a.ImageUpdateImageByNameHandler = image.UpdateImageByNameHandlerFunc(h.updateImage)
}

// DefaultHandlers implements Handlers interface
type defaultHandlers struct {
	backend       backend.Backend
	httpClient    *http.Client
	namespace     string
	imageRegistry string
}

// NewHandlers is the constructor for image manager API Handler
func NewHandlers(kubecfgPath, namespace, imageRegistry string) Handlers {
	return &defaultHandlers{
		backend:       backend.KnativeBuild(kubecfgPath),
		httpClient:    &http.Client{},
		namespace:     namespace,
		imageRegistry: imageRegistry,
	}
}

func (h *defaultHandlers) addImage(params image.AddImageParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	img := params.Body
	utils.AdjustMeta(&img.Meta, dapi.Meta{Name: img.Name, Org: org, Project: project})

	log.Debugf("adding name: %s, org:%s, proj:%s\n", img.Meta.Name, img.Meta.Org, img.Meta.Project)

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

func (h *defaultHandlers) getImage(params image.GetImageByNameParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.ImageName
	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("getting image %s in %s:%s", name, org, project)
	img, err := h.backend.GetImage(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if derrors.IsObjectNotFound(err) {
		log.Debugf("image %s in %s:%s not found", name, org, project)
		return image.NewGetImageByNameNotFound().WithPayload(derrors.GetError(err))
	}
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "get image"))
		return image.NewGetImageByNameDefault(500).WithPayload(derrors.GetError(err))
	}

	return image.NewGetImageByNameOK().WithPayload(img)
}

func (h *defaultHandlers) deleteImage(params image.DeleteImageByNameParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.ImageName
	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("deleting image in %s:%s", org, project)
	err := h.backend.DeleteImage(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if derrors.IsObjectNotFound(err) {
		log.Debugf("image %s in %s:%s not found", name, org, project)
		return image.NewDeleteImageByNameNotFound().WithPayload(derrors.GetError(err))
	}
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "get image"))
		return image.NewDeleteImageByNameDefault(500).WithPayload(derrors.GetError(err))
	}
	return image.NewDeleteImageByNameOK()
}

func (h *defaultHandlers) getImages(params image.GetImagesParams) middleware.Responder {
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

func (h *defaultHandlers) updateImage(params image.UpdateImageByNameParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	img := params.Body
	org := h.namespace
	project := *params.XDispatchProject
	utils.AdjustMeta(&img.Meta, dapi.Meta{Name: img.Name, Org: org, Project: project})

	updated, err := h.backend.UpdateImage(ctx, img)
	if derrors.IsObjectNotFound(err) {
		log.Debugf("image %s in %s:%s not found", img.Name, org, project)
		return image.NewUpdateImageByNameNotFound().WithPayload(derrors.GetError(err))
	}
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "get image"))
		return image.NewUpdateImageByNameDefault(500).WithPayload(derrors.GetError(err))
	}
	return image.NewUpdateImageByNameOK().WithPayload(updated)
}
