///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"fmt"
	"net/http"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/utils"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
	baseimage "github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations/image"
	"github.com/vmware/dispatch/pkg/trace"
)

var statusMap = map[v1.Status]entitystore.Status{
	v1.StatusCREATING:    StatusCREATING,
	v1.StatusUPDATING:    StatusUPDATING,
	v1.StatusDELETED:     StatusDELETED,
	v1.StatusERROR:       StatusERROR,
	v1.StatusINITIALIZED: StatusINITIALIZED,
	v1.StatusREADY:       StatusREADY,
}

var reverseStatusMap = make(map[entitystore.Status]v1.Status)

func initializeStatusMap() {
	for k, v := range statusMap {
		reverseStatusMap[v] = k
	}
}

func baseImageEntityToModel(e *BaseImage) *v1.BaseImage {
	var tags []*v1.Tag
	for k, v := range e.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}

	m := v1.BaseImage{
		CreatedTime: e.CreatedTime.Unix(),
		DockerURL:   swag.String(e.DockerURL),
		Language:    swag.String(e.Language),
		ID:          strfmt.UUID(e.ID),
		Name:        swag.String(e.Name),
		Kind:        utils.BaseImageKind,
		Status:      reverseStatusMap[e.Status],
		Tags:        tags,
		Reason:      e.Reason,
	}
	return &m
}

func baseImageModelToEntity(m *v1.BaseImage) *BaseImage {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := BaseImage{
		BaseEntity: entitystore.BaseEntity{
			Name:   *m.Name,
			Tags:   tags,
			Status: statusMap[m.Status],
			Reason: m.Reason,
		},
		DockerURL: *m.DockerURL,
		Language:  *m.Language,
	}
	return &e
}

func imageEntityToModel(e *Image) *v1.Image {
	var tags []*v1.Tag
	for k, v := range e.Tags {
		tags = append(tags, &v1.Tag{Key: k, Value: v})
	}
	var packages []*v1.SystemDependency
	for i := range e.SystemDependencies.Packages {
		p := e.SystemDependencies.Packages[i]
		packages = append(packages, &v1.SystemDependency{Name: &p.Name, Version: p.Version})
	}
	m := v1.Image{
		CreatedTime:   e.CreatedTime.Unix(),
		BaseImageName: swag.String(e.BaseImageName),
		DockerURL:     e.DockerURL,
		Language:      e.Language,
		RuntimeDependencies: &v1.RuntimeDependencies{
			Manifest: e.RuntimeDependencies.Manifest,
		},
		SystemDependencies: &v1.SystemDependencies{
			Packages: packages,
		},
		ID:     strfmt.UUID(e.ID),
		Name:   swag.String(e.Name),
		Kind:   utils.ImageKind,
		Status: reverseStatusMap[e.Status],
		Tags:   tags,
		Reason: e.Reason,
	}
	return &m
}

func imageModelToEntity(m *v1.Image) *Image {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	var packages []SystemPackage
	if m.SystemDependencies != nil {
		for i := range m.SystemDependencies.Packages {
			p := m.SystemDependencies.Packages[i]
			packages = append(packages, SystemPackage{Name: *p.Name, Version: p.Version})
		}
	}
	var runtimeDeps RuntimeDependencies
	if m.RuntimeDependencies != nil {
		runtimeDeps.Manifest = m.RuntimeDependencies.Manifest
	}
	e := Image{
		BaseEntity: entitystore.BaseEntity{
			Name:   *m.Name,
			Tags:   tags,
			Status: statusMap[m.Status],
			Reason: m.Reason,
		},
		Language:            m.Language,
		BaseImageName:       *m.BaseImageName,
		RuntimeDependencies: runtimeDeps,
		SystemDependencies: SystemDependencies{
			Packages: packages,
		},
	}
	return &e
}

// Handlers encapsulates the image manager handlers
type Handlers struct {
	imageBuilder     *ImageBuilder
	baseImageBuilder *BaseImageBuilder
	Store            entitystore.EntityStore
	Watcher          controller.Watcher
}

// NewHandlers is the constructor for the Handlers type
func NewHandlers(imageBuilder *ImageBuilder, baseImageBuilder *BaseImageBuilder, watcher controller.Watcher, store entitystore.EntityStore) *Handlers {
	return &Handlers{
		imageBuilder:     imageBuilder,
		baseImageBuilder: baseImageBuilder,
		Store:            store,
		Watcher:          watcher,
	}
}

// ConfigureHandlers registers the image manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	a, ok := api.(*operations.ImageManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	initializeStatusMap()

	a.CookieAuth = func(token string) (interface{}, error) {

		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.BaseImageAddBaseImageHandler = baseimage.AddBaseImageHandlerFunc(h.addBaseImage)
	a.BaseImageGetBaseImageByNameHandler = baseimage.GetBaseImageByNameHandlerFunc(h.getBaseImageByName)
	a.BaseImageGetBaseImagesHandler = baseimage.GetBaseImagesHandlerFunc(h.getBaseImages)
	a.BaseImageUpdateBaseImageByNameHandler = baseimage.UpdateBaseImageByNameHandlerFunc(h.updateBaseImageByName)
	a.BaseImageDeleteBaseImageByNameHandler = baseimage.DeleteBaseImageByNameHandlerFunc(h.deleteBaseImageByName)
	a.ImageAddImageHandler = image.AddImageHandlerFunc(h.addImage)
	a.ImageGetImageByNameHandler = image.GetImageByNameHandlerFunc(h.getImageByName)
	a.ImageGetImagesHandler = image.GetImagesHandlerFunc(h.getImages)
	a.ImageUpdateImageByNameHandler = image.UpdateImageByNameHandlerFunc(h.updateImageByName)
	a.ImageDeleteImageByNameHandler = image.DeleteImageByNameHandlerFunc(h.deleteImageByName)
}

func (h *Handlers) addBaseImage(params baseimage.AddBaseImageParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	baseImageRequest := params.Body
	e := baseImageModelToEntity(baseImageRequest)
	e.OrganizationID = params.XDispatchOrg
	e.Status = StatusINITIALIZED
	_, err := h.Store.Add(ctx, e)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return baseimage.NewAddBaseImageConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("base image", e.Name),
			})
		}
		log.Debugf("store error when adding base image: %+v", err)
		return baseimage.NewAddBaseImageDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("base image", e.Name),
			})
	}

	h.Watcher.OnAction(ctx, e)

	m := baseImageEntityToModel(e)
	return baseimage.NewAddBaseImageCreated().WithPayload(m)
}

func (h *Handlers) getBaseImageByName(params baseimage.GetBaseImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := BaseImage{}
	err := h.Store.Get(ctx, params.XDispatchOrg, params.BaseImageName, entitystore.Options{}, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent base image %s", params.BaseImageName)
		log.Debugf("store error when getting base image: %+v", err)
		return baseimage.NewGetBaseImageByNameNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("base image", params.BaseImageName),
			})
	}
	m := baseImageEntityToModel(&e)
	return baseimage.NewGetBaseImageByNameOK().WithPayload(m)
}

func (h *Handlers) getBaseImages(params baseimage.GetBaseImagesParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var images []*BaseImage

	var err error
	opts := entitystore.Options{}
	err = h.Store.List(ctx, params.XDispatchOrg, opts, &images)
	if err != nil {
		log.Errorf("store error when listing base images: %+v", err)
		return baseimage.NewGetBaseImagesDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting base images"),
			})
	}
	var imageModels []*v1.BaseImage
	for _, i := range images {
		imageModels = append(imageModels, baseImageEntityToModel(i))
	}
	return baseimage.NewGetBaseImagesOK().WithPayload(imageModels)
}

func (h *Handlers) updateBaseImageByName(params baseimage.UpdateBaseImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := BaseImage{}
	err := h.Store.Get(ctx, params.XDispatchOrg, params.BaseImageName, entitystore.Options{}, &e)
	if err != nil {
		return baseimage.NewUpdateBaseImageByNameNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("base image", params.BaseImageName),
			})
	}

	baseImageRequest := params.Body
	updateEntity := baseImageModelToEntity(baseImageRequest)

	updateEntity.CreatedTime = e.CreatedTime
	updateEntity.ID = e.ID
	updateEntity.Status = entitystore.StatusUPDATING
	updateEntity.OrganizationID = e.OrganizationID

	_, err = h.Store.Update(ctx, e.Revision, updateEntity)
	if err != nil {
		log.Errorf("store error when updating base image: %+v", err)
		return baseimage.NewUpdateBaseImageByNameDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when updating base image"),
			})
	}

	if err := h.baseImageBuilder.baseImageDelete(ctx, &e); err != nil {
		log.Errorf("error deleting docker image %s: %+v", e.DockerURL, err)
	}

	h.Watcher.OnAction(ctx, updateEntity)

	m := baseImageEntityToModel(updateEntity)
	return baseimage.NewUpdateBaseImageByNameOK().WithPayload(m)
}

func (h *Handlers) deleteBaseImageByName(params baseimage.DeleteBaseImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := BaseImage{}
	err := h.Store.Get(ctx, params.XDispatchOrg, params.BaseImageName, entitystore.Options{}, &e)
	if err != nil {
		return baseimage.NewDeleteBaseImageByNameNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("base image", params.BaseImageName),
			})
	}
	e.Delete = true
	_, err = h.Store.Update(ctx, e.Revision, &e)
	if err != nil {
		log.Errorf("store error when deleting base image: %+v", err)
		return baseimage.NewDeleteBaseImageByNameDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("base image", e.Name),
			})
	}

	h.Watcher.OnAction(ctx, &e)

	m := baseImageEntityToModel(&e)
	return baseimage.NewDeleteBaseImageByNameOK().WithPayload(m)
}

func (h *Handlers) addImage(params image.AddImageParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	imageRequest := params.Body
	e := imageModelToEntity(imageRequest)
	e.OrganizationID = params.XDispatchOrg
	e.Status = StatusINITIALIZED

	var bi BaseImage
	err := h.Store.Get(ctx, e.OrganizationID, e.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		log.Debugf("store error when fetching base image: %+v", err)
		return image.NewAddImageBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Error fetching base image %s", e.BaseImageName)),
			})
	}
	e.Language = bi.Language

	_, err = h.Store.Add(ctx, e)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return image.NewAddImageConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("image", e.Name),
			})
		}
		log.Debugf("store error when adding image: %+v", err)
		return image.NewAddImageDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("image", e.Name),
			})
	}

	h.Watcher.OnAction(ctx, e)

	m := imageEntityToModel(e)
	return image.NewAddImageCreated().WithPayload(m)
}

func (h *Handlers) getImageByName(params image.GetImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := Image{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	err = h.Store.Get(ctx, params.XDispatchOrg, params.ImageName, opts, &e)
	if err != nil {
		log.Warnf("Received GET for non-existentimage %s", params.ImageName)
		log.Debugf("store error when getting image: %+v", err)
		return image.NewGetImageByNameNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("image", params.ImageName),
			})
	}
	m := imageEntityToModel(&e)
	return image.NewGetImageByNameOK().WithPayload(m)
}

func (h *Handlers) getImages(params image.GetImagesParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var images []*Image

	var err error
	opts := entitystore.Options{}

	err = h.Store.List(ctx, params.XDispatchOrg, opts, &images)
	if err != nil {
		log.Errorf("store error when listing images: %+v", err)
		return image.NewGetImagesDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error while listing images"),
			})
	}
	var imageModels []*v1.Image
	for _, i := range images {
		imageModels = append(imageModels, imageEntityToModel(i))
	}

	return image.NewGetImagesOK().WithPayload(imageModels)
}

func (h *Handlers) updateImageByName(params image.UpdateImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	imageRequest := params.Body
	e := imageModelToEntity(imageRequest)
	e.OrganizationID = params.XDispatchOrg

	var bi BaseImage
	err := h.Store.Get(ctx, e.OrganizationID, e.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		log.Debugf("store error when fetching base image: %+v", err)
		return image.NewUpdateImageByNameBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Error fetching base image %s", e.BaseImageName)),
			})
	}
	e.Language = bi.Language

	var current Image
	err = h.Store.Get(ctx, e.OrganizationID, params.ImageName, entitystore.Options{}, &current)
	if err != nil {
		return image.NewUpdateImageByNameNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("image", params.ImageName),
			})
	}

	e.Status = StatusUPDATING
	e.CreatedTime = current.CreatedTime
	e.ID = current.ID
	e.DockerURL = current.DockerURL

	_, err = h.Store.Update(ctx, current.Revision, e)
	if err != nil {
		log.Debugf("store error when updating image: %+v", err)
		return image.NewUpdateImageByNameDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("image", e.Name),
			})
	}

	if h.Watcher != nil {
		h.Watcher.OnAction(ctx, e)
	} else {
		log.Debugf("note: the watcher is nil")
	}

	m := imageEntityToModel(e)
	return image.NewUpdateImageByNameOK().WithPayload(m)
}

func (h *Handlers) deleteImageByName(params image.DeleteImageByNameParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	e := Image{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	err = h.Store.Get(ctx, params.XDispatchOrg, params.ImageName, opts, &e)
	if err != nil {
		return image.NewDeleteImageByNameNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("image", params.ImageName),
			})
	}
	e.Delete = true
	_, err = h.Store.Update(ctx, e.Revision, &e)
	if err != nil {
		log.Errorf("store error when deleting image: %+v", err)
		return image.NewDeleteImageByNameDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("image", e.Name),
			})
	}
	e.Status = StatusDELETED

	h.Watcher.OnAction(ctx, &e)

	m := imageEntityToModel(&e)
	return image.NewDeleteImageByNameOK().WithPayload(m)
}
