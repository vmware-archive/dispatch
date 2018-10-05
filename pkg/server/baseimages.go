///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"

	baseimages "github.com/vmware/dispatch/pkg/baseimages"
	"github.com/vmware/dispatch/pkg/baseimages/gen/restapi"
	"github.com/vmware/dispatch/pkg/baseimages/gen/restapi/operations"
)

func initBaseImages(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewBaseImagesAPI(swaggerSpec)
	handlers := baseimages.NewHandlers(config.K8sConfig, config.Namespace)

	baseimages.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
