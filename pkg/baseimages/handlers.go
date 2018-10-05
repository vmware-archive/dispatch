///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package baseimages

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/baseimages/backend"
	"github.com/vmware/dispatch/pkg/baseimages/gen/restapi/operations"
	baseimage "github.com/vmware/dispatch/pkg/baseimages/gen/restapi/operations/base_image"
	derrors "github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Handlers interface declares methods for image-manage API
// pricinpal interface{} reserved for security authentication
type Handlers interface {
	addBaseImage(params baseimage.AddBaseImageParams, principal interface{}) middleware.Responder
	getBaseImage(params baseimage.GetBaseImageByNameParams, principal interface{}) middleware.Responder
	deleteBaseImage(params baseimage.DeleteBaseImageByNameParams, principal interface{}) middleware.Responder
	getBaseImages(params baseimage.GetBaseImagesParams, principal interface{}) middleware.Responder
	updateBaseImage(params baseimage.UpdateBaseImageByNameParams, principal interface{}) middleware.Responder
}

// ConfigureHandlers registers the image manager handlers to API
func ConfigureHandlers(api middleware.RoutableAPI, h Handlers) {
	a, ok := api.(*operations.BaseImagesAPI)
	if !ok {
		panic("Cannot configure image manager apis")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.Logger = log.Printf

	a.BaseImageAddBaseImageHandler = baseimage.AddBaseImageHandlerFunc(h.addBaseImage)
	a.BaseImageGetBaseImageByNameHandler = baseimage.GetBaseImageByNameHandlerFunc(h.getBaseImage)
	a.BaseImageDeleteBaseImageByNameHandler = baseimage.DeleteBaseImageByNameHandlerFunc(h.deleteBaseImage)
	a.BaseImageGetBaseImagesHandler = baseimage.GetBaseImagesHandlerFunc(h.getBaseImages)
	a.BaseImageUpdateBaseImageByNameHandler = baseimage.UpdateBaseImageByNameHandlerFunc(h.updateBaseImage)
}

// DefaultHandlers implements Handlers interface
type defaultHandlers struct {
	backend       backend.Backend
	httpClient    *http.Client
	namespace     string
	imageRegistry string
}

// NewHandlers is the constructor for image manager API Handler
func NewHandlers(kubecfgPath, namespace string) Handlers {
	return &defaultHandlers{
		backend:   backend.KnativeBuild(kubecfgPath),
		namespace: namespace,
	}
}

func (h *defaultHandlers) addBaseImage(params baseimage.AddBaseImageParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	model := params.Body
	utils.AdjustMeta(&model.Meta, dapi.Meta{Name: model.Name, Org: org, Project: project})

	log.Debugf("adding name: %s, org:%s, proj:%s\n", model.Meta.Name, model.Meta.Org, model.Meta.Project)

	createdBaseImage, err := h.backend.AddBaseImage(ctx, model)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "creating a base image"))
		return baseimage.NewAddBaseImageDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("baseimage", model.Meta.Name),
		})
	}
	return baseimage.NewAddBaseImageCreated().WithPayload(createdBaseImage)
}

func (h *defaultHandlers) getBaseImage(params baseimage.GetBaseImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.BaseImageName
	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("getting baseimage %s in %s:%s", name, org, project)
	img, err := h.backend.GetBaseImage(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if derrors.IsObjectNotFound(err) {
			log.Debugf("baseimage %s in %s:%s not found", name, org, project)
			return baseimage.NewGetBaseImageByNameNotFound().WithPayload(derrors.GetError(err))
		}
		log.Errorf("%+v", errors.Wrap(err, "get image"))
		return baseimage.NewGetBaseImageByNameDefault(500).WithPayload(derrors.GetError(err))
	}
	return baseimage.NewGetBaseImageByNameOK().WithPayload(img)
}

func (h *defaultHandlers) deleteBaseImage(params baseimage.DeleteBaseImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.BaseImageName
	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("deleting baseimage in %s:%s", org, project)
	err := h.backend.DeleteBaseImage(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if derrors.IsObjectNotFound(err) {
			log.Debugf("baseimage %s in %s:%s not found", name, org, project)
			return baseimage.NewDeleteBaseImageByNameNotFound().WithPayload(derrors.GetError(err))
		}
		log.Errorf("%+v", errors.Wrap(err, "get baseimage"))
		return baseimage.NewDeleteBaseImageByNameDefault(500).WithPayload(derrors.GetError(err))
	}
	return baseimage.NewDeleteBaseImageByNameOK()
}

func (h *defaultHandlers) getBaseImages(params baseimage.GetBaseImagesParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("getting baseimages in %s:%s", org, project)
	dImages, err := h.backend.ListBaseImage(ctx, &dapi.Meta{Org: org, Project: project})
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "listing baseimages"))
		return baseimage.NewGetBaseImagesDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String(err.Error()),
		})
	}

	return baseimage.NewGetBaseImagesOK().WithPayload(dImages)
}

func (h *defaultHandlers) updateBaseImage(params baseimage.UpdateBaseImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	model := params.Body
	org := h.namespace
	project := *params.XDispatchProject
	utils.AdjustMeta(&model.Meta, dapi.Meta{Name: model.Name, Org: org, Project: project})

	updated, err := h.backend.UpdateBaseImage(ctx, model)
	if err != nil {
		if derrors.IsObjectNotFound(err) {
			log.Debugf("baseimage %s in %s:%s not found", model.Name, org, project)
			return baseimage.NewUpdateBaseImageByNameNotFound().WithPayload(derrors.GetError(err))
		}
		log.Errorf("%+v", errors.Wrap(err, "get baseimage"))
		return baseimage.NewUpdateBaseImageByNameDefault(500).WithPayload(derrors.GetError(err))
	}
	return baseimage.NewUpdateBaseImageByNameOK().WithPayload(updated)
}
