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
	DbFile string `long:"db-file" description:"Path to BoltDB file" default:"db.bolt"`
	OrgID  string `long:"organization" description:"(temporary) Static organization id" default:"serverless"`
}{}

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
		Status:      models.StatusREADY,
		Tags:        tags,
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
		},
		DockerURL: *m.DockerURL,
		Public:    *m.Public,
	}
	return &e
}

// ConfigureHandlers registers the image manager handlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, store entitystore.EntityStore) {

	a, ok := api.(*operations.ImageManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.BaseImageAddBaseImageHandler = baseimage.AddBaseImageHandlerFunc(func(params baseimage.AddBaseImageParams) middleware.Responder {
		baseImageRequest := params.Body
		e := baseImageModelToEntity(baseImageRequest)
		_, err := store.Add(e)
		if err != nil {
			return baseimage.NewAddBaseImageBadRequest()
		}
		m := baseImageEntityToModel(e)
		return baseimage.NewAddBaseImageCreated().WithPayload(m)
	})

	a.BaseImageGetBaseImageByIDHandler = baseimage.GetBaseImageByIDHandlerFunc(func(params baseimage.GetBaseImageByIDParams) middleware.Responder {
		e := BaseImage{}
		err := store.GetById(ImageManagerFlags.OrgID, params.BaseImageID.String(), &e)
		if err != nil {
			return baseimage.NewGetBaseImageByIDNotFound()
		}
		m := baseImageEntityToModel(&e)
		return baseimage.NewGetBaseImageByIDOK().WithPayload(m)
	})
}
