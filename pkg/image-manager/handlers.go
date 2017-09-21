///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package imagemanager

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations"
	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/restapi/operations/base_image"
)

// ImageManagerFlags are configuration flags for the image manager
var ImageManagerFlags = struct {
	DbFile       string `long:"db-file" description:"Path to BoltDB file" default:"./db.bolt"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"serverless"`
	K8sConfig    string `long:"kubeconfig" description:"Path to kubernetes config file"`
	K8sNamespace string `long:"namespace" description:"Kubernetes namespace for jobs" default:"default"`
}{}

var statusMap = map[models.Status]entitystore.Status{
	models.StatusCREATING:    StatusCREATING,
	models.StatusDELETED:     StatusDELETED,
	models.StatusERROR:       StatusERROR,
	models.StatusINITIALIZED: StatusINITIALIZED,
	models.StatusREADY:       StatusREADY,
}

var reverseStatusMap = make(map[entitystore.Status]models.Status)

func initializeStatusMap() {
	for k, v := range statusMap {
		reverseStatusMap[v] = k
	}
}

func baseImageEntityToModel(e *BaseImage) *models.BaseImage {
	var tags []*models.Tag
	for k, v := range e.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}

	m := models.BaseImage{
		CreatedTime: e.CreatedTime.Unix(),
		DockerURL:   swag.String(e.DockerURL),
		ID:          strfmt.UUID(e.ID),
		Public:      swag.Bool(e.Public),
		Name:        swag.String(e.Name),
		Status:      reverseStatusMap[e.Status],
		Tags:        tags,
		Reason:      e.Reason,
	}
	return &m
}

func baseImageModelToEntity(m *models.BaseImage) *BaseImage {
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
		Public:    *m.Public,
	}
	return &e
}

// Handlers encapsulates the image manager handlers
type Handlers struct {
	baseImageBuilder *BaseImageBuilder
}

// NewHandlers is the constructor for the Handlers type
func NewHandlers(baseImageBuilder *BaseImageBuilder) *Handlers {
	return &Handlers{
		baseImageBuilder: baseImageBuilder,
	}
}

// ConfigureHandlers registers the image manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI, store entitystore.EntityStore) {

	a, ok := api.(*operations.ImageManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	initializeStatusMap()

	a.BaseImageAddBaseImageHandler = baseimage.AddBaseImageHandlerFunc(func(params baseimage.AddBaseImageParams) middleware.Responder {
		baseImageRequest := params.Body
		e := baseImageModelToEntity(baseImageRequest)
		e.Status = StatusINITIALIZED
		_, err := store.Add(e)
		if err != nil {
			return baseimage.NewAddBaseImageBadRequest()
		}
		if h.baseImageBuilder != nil {
			h.baseImageBuilder.baseImageChannel <- *e
		}
		m := baseImageEntityToModel(e)
		return baseimage.NewAddBaseImageCreated().WithPayload(m)
	})

	a.BaseImageGetBaseImageByNameHandler = baseimage.GetBaseImageByNameHandlerFunc(func(params baseimage.GetBaseImageByNameParams) middleware.Responder {
		e := BaseImage{}
		err := store.Get(ImageManagerFlags.OrgID, params.BaseImageName, &e)
		if err != nil {
			return baseimage.NewGetBaseImageByNameNotFound()
		}
		m := baseImageEntityToModel(&e)
		return baseimage.NewGetBaseImageByNameOK().WithPayload(m)
	})

	a.BaseImageGetBaseImagesHandler = baseimage.GetBaseImagesHandlerFunc(func(params baseimage.GetBaseImagesParams) middleware.Responder {
		var images []BaseImage
		err := store.List(ImageManagerFlags.OrgID, nil, &images)
		if err != nil {
			return baseimage.NewGetBaseImagesDefault(500)
		}
		var imageModels []*models.BaseImage
		for _, image := range images {
			imageModels = append(imageModels, baseImageEntityToModel(&image))
		}
		return baseimage.NewGetBaseImagesOK().WithPayload(imageModels)
	})

	a.ServerShutdown = func() {
		h.baseImageBuilder.done <- true
	}
}
