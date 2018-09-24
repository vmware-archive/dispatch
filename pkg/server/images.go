///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"

	images "github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/images/gen/restapi"
	"github.com/vmware/dispatch/pkg/images/gen/restapi/operations"
)

func initImages(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewImagesAPI(swaggerSpec)
	handlers := images.NewHandlers(config.K8sConfig, config.Namespace, config.ImageRegistry)

	images.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
