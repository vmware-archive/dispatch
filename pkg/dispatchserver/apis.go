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
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/api-manager/istio"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

type apisConfig struct {
	// API Manager config option
	GatewayHost string `mapstructure:"gateway-host" json:"gateway-host,omitempty"`
}

// NewCmdAPIs creates a subcommand to run api manager
func NewCmdAPIs(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "api-manager",
		Short:  i18n.T("Run Dispatch API Manager"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.APIs),
		Run: func(cmd *cobra.Command, args []string) {
			runAPIs(config)
		},
	}
	cmd.SetOutput(out)

	cmd.Flags().String("gateway-host", "gateway-kong", "Admin Endpoint for API Gateway backend.")

	return cmd
}

func runAPIs(config *serverConfig) {
	apisHandler := initAPIs(config)

	handler := addMiddleware(apisHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initAPIs(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewAPIManagerAPI(swaggerSpec)

	handlers := apimanager.NewHandlers()

	client, err := istio.NewClient()
	if err != nil {
		log.Fatalf("Unable to connect an istio client: %v", err)
	}

	istioHandlers := apimanager.NewIstioHandlers(client)

	handlers.ConfigureHandlers(api, istioHandlers)

	return api.Serve(nil)
}
