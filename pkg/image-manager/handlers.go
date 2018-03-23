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

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager/gen/models"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
	baseimage "github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations/image"
	"github.com/vmware/dispatch/pkg/trace"
)

// ImageManagerFlags are configuration flags for the image manager
var ImageManagerFlags = struct {
	Config       string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile       string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend    string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser       string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword   string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase   string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ResyncPeriod int    `long:"resync-period" description:"The time period (in seconds) to sync with image repository" default:"10"`
}{}

var statusMap = map[models.Status]entitystore.Status{
	models.StatusCREATING:    StatusCREATING,
	models.StatusUPDATING:    StatusUPDATING,
	models.StatusDELETED:     StatusDELETED,
	models.StatusERROR:       StatusERROR,
	models.StatusINITIALIZED: StatusINITIALIZED,
	models.StatusREADY:       StatusREADY,
}

var reverseStatusMap = make(map[entitystore.Status]models.Status)

func initializeStatusMap() {
	defer trace.Trace("initializeStatusMap")()
	for k, v := range statusMap {
		reverseStatusMap[v] = k
	}
}

func baseImageEntityToModel(e *BaseImage) *models.BaseImage {
	defer trace.Trace("baseImageEntityToModel")()
	var tags []*models.Tag
	for k, v := range e.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}

	m := models.BaseImage{
		CreatedTime: e.CreatedTime.Unix(),
		DockerURL:   swag.String(e.DockerURL),
		Language:    models.Language(e.Language),
		ID:          strfmt.UUID(e.ID),
		Name:        swag.String(e.Name),
		Kind:        utils.BaseImageKind,
		Status:      reverseStatusMap[e.Status],
		Tags:        tags,
		Reason:      e.Reason,
	}
	return &m
}

func baseImageModelToEntity(m *models.BaseImage) *BaseImage {
	defer trace.Trace("baseImageModelToEntity")()
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := BaseImage{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: ImageManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
			Status:         statusMap[m.Status],
			Reason:         m.Reason,
		},
		DockerURL: *m.DockerURL,
		Language:  Language(string(m.Language)),
	}
	return &e
}

func imageEntityToModel(e *Image) *models.Image {
	defer trace.Trace("imageEntityToModel")()
	var tags []*models.Tag
	for k, v := range e.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	var packages []*models.SystemDependency
	for i := range e.SystemDependencies.Packages {
		p := e.SystemDependencies.Packages[i]
		packages = append(packages, &models.SystemDependency{Name: &p.Name, Version: p.Version})
	}
	m := models.Image{
		CreatedTime:   e.CreatedTime.Unix(),
		BaseImageName: swag.String(e.BaseImageName),
		DockerURL:     e.DockerURL,
		Language:      models.Language(e.Language),
		RuntimeDependencies: &models.RuntimeDependencies{
			Format:   e.RuntimeDependencies.Format,
			Manifest: e.RuntimeDependencies.Manifest,
		},
		SystemDependencies: &models.SystemDependencies{
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

func imageModelToEntity(m *models.Image) *Image {
	defer trace.Trace("imageModelToEntity")()
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
		runtimeDeps.Format = m.RuntimeDependencies.Format
		runtimeDeps.Manifest = m.RuntimeDependencies.Manifest
	}
	e := Image{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: ImageManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
			Status:         statusMap[m.Status],
			Reason:         m.Reason,
		},
		DockerURL:           m.DockerURL,
		Language:            Language(string(m.Language)),
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
	defer trace.Trace("NewHandlers")()
	return &Handlers{
		imageBuilder:     imageBuilder,
		baseImageBuilder: baseImageBuilder,
		Store:            store,
		Watcher:          watcher,
	}
}

// ConfigureHandlers registers the image manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	defer trace.Trace("ConfigureHandlers")()
	a, ok := api.(*operations.ImageManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	initializeStatusMap()

	a.CookieAuth = func(token string) (interface{}, error) {

		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		log.Printf("cookie auth: %s\n", token)
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
	defer trace.Trace("addBaseImage")()
	baseImageRequest := params.Body
	e := baseImageModelToEntity(baseImageRequest)
	e.Status = StatusINITIALIZED
	_, err := h.Store.Add(e)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return baseimage.NewAddBaseImageConflict().WithPayload(&models.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating base image: non-unique name"),
			})
		}
		log.Debugf("store error when adding base image: %+v", err)
		return baseimage.NewAddBaseImageBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String("store error when adding base image"),
			})
	}

	h.Watcher.OnAction(e)

	m := baseImageEntityToModel(e)
	return baseimage.NewAddBaseImageCreated().WithPayload(m)
}

func (h *Handlers) getBaseImageByName(params baseimage.GetBaseImageByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getBaseImageByName")()
	e := BaseImage{}
	err := h.Store.Get(ImageManagerFlags.OrgID, params.BaseImageName, entitystore.Options{}, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent base image %s", params.BaseImageName)
		log.Debugf("store error when getting base image: %+v", err)
		return baseimage.NewGetBaseImageByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("base image %s not found", params.BaseImageName)),
			})
	}
	m := baseImageEntityToModel(&e)
	return baseimage.NewGetBaseImageByNameOK().WithPayload(m)
}

func (h *Handlers) getBaseImages(params baseimage.GetBaseImagesParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getBaseImages")()
	var images []*BaseImage

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err = h.Store.List(ImageManagerFlags.OrgID, opts, &images)
	if err != nil {
		log.Errorf("store error when listing base images: %+v", err)
		return baseimage.NewGetBaseImagesDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting base images"),
			})
	}
	var imageModels []*models.BaseImage
	for _, image := range images {
		imageModels = append(imageModels, baseImageEntityToModel(image))
	}
	return baseimage.NewGetBaseImagesOK().WithPayload(imageModels)
}

func (h *Handlers) updateBaseImageByName(params baseimage.UpdateBaseImageByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("updateBaseImageByName")()
	e := BaseImage{}
	err := h.Store.Get(ImageManagerFlags.OrgID, params.BaseImageName, entitystore.Options{}, &e)
	if err != nil {
		return baseimage.NewUpdateBaseImageByNameNotFound()
	}

	baseImageRequest := params.Body
	updateEntity := baseImageModelToEntity(baseImageRequest)

	updateEntity.CreatedTime = e.CreatedTime
	updateEntity.ID = e.ID
	updateEntity.Status = entitystore.StatusUPDATING

	_, err = h.Store.Update(e.Revision, updateEntity)
	if err != nil {
		log.Errorf("store error when updating base image: %+v", err)
		return baseimage.NewUpdateBaseImageByNameDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when updating base image"),
			})
	}

	h.Watcher.OnAction(updateEntity)

	m := baseImageEntityToModel(updateEntity)
	return baseimage.NewUpdateBaseImageByNameOK().WithPayload(m)
}

func (h *Handlers) deleteBaseImageByName(params baseimage.DeleteBaseImageByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("deleteBaseImageByName")()
	e := BaseImage{}
	err := h.Store.Get(ImageManagerFlags.OrgID, params.BaseImageName, entitystore.Options{}, &e)
	if err != nil {
		return baseimage.NewDeleteBaseImageByNameNotFound()
	}
	e.Delete = true
	_, err = h.Store.Update(e.Revision, &e)
	if err != nil {
		log.Errorf("store error when deleting base image: %+v", err)
		return baseimage.NewDeleteBaseImageByNameDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when deleting base image"),
			})
	}

	h.Watcher.OnAction(&e)

	m := baseImageEntityToModel(&e)
	return baseimage.NewDeleteBaseImageByNameOK().WithPayload(m)
}

func (h *Handlers) addImage(params image.AddImageParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	imageRequest := params.Body
	e := imageModelToEntity(imageRequest)
	e.Status = StatusINITIALIZED

	var bi BaseImage
	err := h.Store.Get(e.OrganizationID, e.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		log.Debugf("store error when fetching base image: %+v", err)
		return image.NewAddImageBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Error fetching base image %s", e.BaseImageName)),
			})
	}
	e.Language = bi.Language

	_, err = h.Store.Add(e)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return image.NewAddImageConflict().WithPayload(&models.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating image: non-unique name"),
			})
		}
		log.Debugf("store error when adding image: %+v", err)
		return image.NewAddImageBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String("store error when adding image"),
			})
	}

	h.Watcher.OnAction(e)

	m := imageEntityToModel(e)
	return image.NewAddImageCreated().WithPayload(m)
}

func (h *Handlers) getImageByName(params image.GetImageByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getImageByName")()
	e := Image{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return image.NewGetImageByNameBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.Store.Get(ImageManagerFlags.OrgID, params.ImageName, opts, &e)
	if err != nil {
		log.Warnf("Received GET for non-existentimage %s", params.ImageName)
		log.Debugf("store error when getting image: %+v", err)
		return image.NewGetImageByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("image %s not found", params.ImageName)),
			})
	}
	m := imageEntityToModel(&e)
	return image.NewGetImageByNameOK().WithPayload(m)
}

func (h *Handlers) getImages(params image.GetImagesParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getImages")()
	var images []*Image

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return image.NewGetImagesBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	err = h.Store.List(ImageManagerFlags.OrgID, opts, &images)
	if err != nil {
		log.Errorf("store error when listing images: %+v", err)
		return image.NewGetImagesDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error while listing images"),
			})
	}
	var imageModels []*models.Image
	for _, image := range images {
		imageModels = append(imageModels, imageEntityToModel(image))
	}

	return image.NewGetImagesOK().WithPayload(imageModels)
}

func (h *Handlers) updateImageByName(params image.UpdateImageByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("updateImageByName")()

	imageRequest := params.Body
	e := imageModelToEntity(imageRequest)

	var bi BaseImage
	err := h.Store.Get(e.OrganizationID, e.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		log.Debugf("store error when fetching base image: %+v", err)
		return image.NewUpdateImageByNameBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Error fetching base image %s", e.BaseImageName)),
			})
	}
	e.Language = bi.Language

	var current Image
	err = h.Store.Get(e.OrganizationID, params.ImageName, entitystore.Options{}, &current)
	if err != nil {
		return image.NewUpdateImageByNameBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Error fetching image %s", params.ImageName)),
			})
	}

	e.Status = StatusUPDATING
	e.CreatedTime = current.CreatedTime
	e.ID = current.ID

	_, err = h.Store.Update(current.Revision, e)
	if err != nil {
		log.Debugf("store error when updating image: %+v", err)
		return image.NewUpdateImageByNameBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String("store error when updating image"),
			})
	}

	if h.Watcher != nil {
		h.Watcher.OnAction(e)
	} else {
		log.Debugf("note: the watcher is nil")
	}

	m := imageEntityToModel(e)
	return image.NewUpdateImageByNameOK().WithPayload(m)
}

func (h *Handlers) deleteImageByName(params image.DeleteImageByNameParams, principal interface{}) middleware.Responder {
	defer trace.Trace("deleteImageByName")()
	e := Image{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return image.NewDeleteImageByNameBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.Store.Get(ImageManagerFlags.OrgID, params.ImageName, opts, &e)
	if err != nil {
		return image.NewDeleteImageByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("image not found"),
			})
	}
	err = h.Store.Delete(ImageManagerFlags.OrgID, params.ImageName, &Image{})
	if err != nil {
		return image.NewDeleteImageByNameNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("image not found while deleting"),
			})
	}
	e.Delete = true
	e.Status = StatusDELETED

	h.Watcher.OnAction(&e)

	m := imageEntityToModel(&e)
	return image.NewDeleteImageByNameOK().WithPayload(m)
}
