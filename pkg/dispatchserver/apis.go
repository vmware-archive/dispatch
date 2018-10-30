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
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/entity-store"
)

type apisConfig struct {
	// API Manager config option
	GatewayHost string `mapstructure:"gateway-host" json:"gateway-host,omitempty"`
}

func initAPIs(config *serverConfig, store entitystore.EntityStore, gw gateway.Gateway) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewAPIManagerAPI(swaggerSpec)

	apiController := apimanager.NewController(&apimanager.ControllerConfig{
		ResyncPeriod: config.ResyncPeriod,
	}, store, gw)
	apiController.Start()

	handlers := apimanager.NewHandlers(apiController.Watcher(), store)

	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		apiController.Shutdown()
	}
}
