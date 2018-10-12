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
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
	"github.com/vmware/dispatch/pkg/events/transport"
	dockerfaas "github.com/vmware/dispatch/pkg/functions/docker"
	"github.com/vmware/dispatch/pkg/http"
	"github.com/vmware/dispatch/pkg/secret-store/service"
)

type localConfig struct {
	DockerHost     string `mapstructure:"docker-host" json:"docker-host,omitempty"`
	GatewayPort    int    `mapstructure:"gateway-port" json:"gateway-port,omitempty"`
	GatewayTLSPort int    `mapstructure:"gateway-tls-port" json:"gateway-tls-port,omitempty"`
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

	cmd.Flags().String("docker-host", "127.0.0.1", "Docker host/IP. It must be reachable from Dispatch Server.")
	cmd.Flags().Int("gateway-port", 8081, "Port for local API Gateway")
	cmd.Flags().Int("gateway-tls-port", 8444, "TLS port for local API Gateway (only when TLS Enabled in global flags)")

	return cmd
}

func runLocal(config *serverConfig) {
	config.DisableRegistry = true

	store := entityStore(config)
	docker := dockerClient(config)
	functions := functionsClient(config)
	secrets := secretsClient(config)
	services := servicesClient(config)
	images := imagesClient(config)

	secretsService := &service.DBSecretsService{EntityStore: store}
	secretsHandler := initSecrets(config, secretsService)

	imagesHandler, imagesShutdown := initImages(config, store)
	defer imagesShutdown()

	faas := dockerfaas.New(docker)
	functionsDeps := functionsDependencies{
		store:          store,
		faas:           faas,
		dockerclient:   docker,
		imagesClient:   images,
		secretsClient:  secrets,
		servicesClient: services,
	}
	functionsHandler, functionsShutdown := initFunctions(config, functionsDeps)
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

	eventTransport := transport.NewInMemory()
	eventBackend, err := drivers.NewDockerBackend(secrets)
	if err != nil {
		log.Errorf("Error creating docker backend for driver: %v", err)
	}

	eventsDeps := eventsDependencies{
		store:     store,
		transport: eventTransport,
		// IN_PROGRESS: add docker as driver backend
		driversBackend:  eventBackend,
		functionsClient: functions,
		secretsClient:   secrets,
	}
	eventsHandler, eventsShutdown := initEvents(config, eventsDeps)
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
