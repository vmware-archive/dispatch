///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/endpoints"
	"github.com/vmware/dispatch/pkg/endpoints/gen/restapi"
	"github.com/vmware/dispatch/pkg/endpoints/gen/restapi/operations"
)

func initEndpoints(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewEndpointsAPI(swaggerSpec)
	handlers := endpoints.NewHandlers(
		config.K8sConfig, config.Namespace, config.InternalGateway,
		config.SharedGateway, config.DispatchHost)
	endpoints.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
