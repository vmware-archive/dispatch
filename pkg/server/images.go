///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	apiclient "github.com/go-openapi/runtime/client"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
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
	// TODO: address dummy auth
	auth := apiclient.APIKeyAuth("cookie", "header", "UNSET")
	baseImagesClient := client.NewBaseImagesClient(fmt.Sprintf("localhost:%d", config.Port), auth, config.Namespace)
	handlers := images.NewHandlers(config.K8sConfig, config.Namespace, config.ImageRegistry, baseImagesClient)

	images.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
