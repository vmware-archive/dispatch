///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api-manager/gateway/local"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/http"
)

type localServer struct {
	DockerHost     string `mapstructure:"docker-host" json:"docker-host"`
	GatewayPort    int    `mapstructure:"gateway-port" json:"gateway-port"`
	GatewayTLSPort int    `mapstructure:"gateway-tls-port" json:"gateway-tls-port"`
}

// NewCmdLocal creates a subcommand to run Dispatch Local server
func NewCmdLocal(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "local",
		Short:  i18n.T("Run Dispatch local server with all services"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Local),
		Run: func(cmd *cobra.Command, args []string) {
			runLocal(config)
		},
	}
	cmd.SetOutput(out)

	cmd.LocalFlags().String("docker-host", "127.0.0.1", "Docker host/IP. It must be reachable from Dispatch Server.")
	cmd.LocalFlags().Int("gateway-port", 8081, "Port for local API Gateway")
	cmd.LocalFlags().Int("gateway-tls-port", 8444, "TLS port for local API Gateway (only when TLS Enabled in global flags)")

	return cmd
}

func runLocal(config *serverConfig) {
	store := entityStore(config)
	docker := dockerClient(config)
	functions := functionsClient(config)
	secrets := secretsClient(config)
	services := servicesClient(config)
	images := imagesClient(config)

	secretsHandler := initSecrets(config, store)

	imagesHandler, imagesShutdown := initImages(config, store)
	defer imagesShutdown()

	functionsHandler, functionsShutdown := initFunctions(config, store, docker, images, secrets, services)
	defer functionsShutdown()

	gw, err := local.NewGateway(functions)
	if err != nil {
		log.Fatalf("Error creating API Gateway: %v", err)
	}

	gw.Server = httpServer(config)
	gw.Server.Port = config.Local.GatewayPort
	gw.Server.TLSPort = config.Local.GatewayTLSPort
	gw.Server.Name = "API Gateway"
	go func() {
		err := gw.Serve()
		if err != nil {
			log.Errorf("Error running API Gateway: %v", err)
		}
	}()

	apisHandler, apisShutdown := initAPIs(config, store, gw)
	defer apisShutdown()

	eventsHandler, eventsShutdown := initEvents(config, store, functions, secrets)
	defer eventsShutdown()

	dispatchHandler := &http.AllInOneRouter{
		FunctionsHandler: functionsHandler,
		ImagesHandler:    imagesHandler,
		SecretsHandler:   secretsHandler,
		EventsHandler:    eventsHandler,
		APIHandler:       apisHandler,
	}
	handler := addMiddleware(dispatchHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func getLocalEndpoint(config *serverConfig) string {
	if !config.DisableHTTP {
		return fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	}
	return fmt.Sprintf("https://%s:%d", config.Host, config.TLSPort)
}
