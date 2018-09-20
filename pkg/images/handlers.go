///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package images

import (
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/images/gen/restapi/operations"
	image "github.com/vmware/dispatch/pkg/images/gen/restapi/operations/image"
)

// Handlers interface declares methods for image-manage API
// pricinpal interface{} reserved for security authentication
type Handlers interface {
	// addBaseImage(params baseimage.AddBaseImageParams, principal interface{}) middleware.Responder
	// getBaseImage(params baseimage.GetBaseImageByNameParams, principal interface{}) middleware.Responder
	// deleteBaseImage(params baseimage.DeleteBaseImageByNameParams, principal interface{}) middleware.Responder
	// getBaseImages(params baseimage.GetBaseImagesParams, principal interface{}) middleware.Responder
	// updateBaseImage(params baseimage.UpdateBaseImageByNameParams, principal interface{}) middleware.Responder

	addImage(params image.AddImageParams, principal interface{}) middleware.Responder
	// getImage(params image.GetImageByNameParams, principal interface{}) middleware.Responder
	// deleteImage(params image.DeleteImageByNameParams, principal interface{}) middleware.Responder
	getImages(params image.GetImagesParams, principal interface{}) middleware.Responder
	// updateImage(params image.UpdateImageByNameParams, principal interface{}) middleware.Responder
}

// ConfigureHandlers registers the image manager handlers to API
func ConfigureHandlers(api middleware.RoutableAPI, h Handlers) {
	a, ok := api.(*operations.ImagesAPI)
	if !ok {
		panic("Cannot configure image manager apis")
	}

	// TODO: authentication CookieAuth/BearerAuth

	a.Logger = log.Printf

	// a.BaseImageAddBaseImageHandler = baseimage.AddBaseImageHandlerFunc(h.addBaseImage)
	// a.BaseImageGetBaseImageByNameHandler = baseimage.GetBaseImageByNameHandlerFunc(h.getBaseImage)
	// a.BaseImageDeleteBaseImageByNameHandler = baseimage.DeleteBaseImageByNameHandlerFunc(h.deleteBaseImage)
	// a.BaseImageGetBaseImagesHandler = baseimage.GetBaseImagesHandlerFunc(h.getBaseImages)
	// a.BaseImageUpdateBaseImageByNameHandler = baseimage.UpdateBaseImageByNameHandlerFunc(h.updateBaseImage)

	a.ImageAddImageHandler = image.AddImageHandlerFunc(h.addImage)
	// a.ImageGetImageByNameHandler = image.GetImageByNameHandlerFunc(h.getImage)
	// a.ImageDeleteImageByNameHandler = image.DeleteImageByNameHandlerFunc(h.deleteImage)
	a.ImageGetImagesHandler = image.GetImagesHandlerFunc(h.getImages)
	// a.ImageUpdateImageByNameHandler = image.UpdateImageByNameHandlerFunc(h.updateImage)
}
