///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/api-manager"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
)

func initAPIs(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewAPIManagerAPI(swaggerSpec)

	handlers := apimanager.NewHandlers()

	istioHandlers := apimanager.NewIstioHandlers()

	handlers.ConfigureHandlers(api, istioHandlers)

	return api.Serve(nil)
}
