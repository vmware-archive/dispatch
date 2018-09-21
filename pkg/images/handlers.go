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
