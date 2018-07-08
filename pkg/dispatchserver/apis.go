///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api-manager"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api-manager/gateway/kong"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/entity-store"
)

// NewCmdAPIs creates a subcommand to run api manager
func NewCmdAPIs(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apis",
		Short: i18n.T("Run Dispatch API Manager"),
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runAPIs(config)
		},
	}
	cmd.SetOutput(out)
	return cmd
}

func runAPIs(config *serverConfig) {
	store := entityStore(config)

	gw, err := kong.NewClient(nil)
	if err != nil {
		log.Fatalf("Error creating Kong client: %v", err)
	}

	apisHandler, shutdown := initAPIs(config, store, gw)
	defer shutdown()

	handler := addMiddleware(apisHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err = server.Serve(); err != nil {
		log.Error(err)
	}
}

func initAPIs(config *serverConfig, store entitystore.EntityStore, gw gateway.Gateway) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewAPIManagerAPI(swaggerSpec)

	apiController := apimanager.NewController(&apimanager.ControllerConfig{ResyncPeriod: config.ResyncPeriod}, store, gw)
	apiController.Start()

	handlers := apimanager.NewHandlers(apiController.Watcher(), store)

	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		apiController.Shutdown()
	}
}
