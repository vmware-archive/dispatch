///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

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
	APIEndpoint    string `mapstructure:"api-endpoint" json:api-endpoint,omitempty"`
}

var dispatchConfigPath = ""

// NewCLI creates cobra object for top-level Dispatch server
func NewCLI(out io.Writer) *cobra.Command {
	log.SetOutput(out)
	cmd := &cobra.Command{
		Use:    "dispatch-server",
		Short:  i18n.T("Dispatch is a batteries-included serverless framework."),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&defaultConfig.Local),
		Run: func(cmd *cobra.Command, args []string) {
			runLocal(defaultConfig)
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig(cmd, defaultConfig)
		},
	}
	cmd.SetOutput(out)

	configGlobalFlags(cmd.PersistentFlags())
	cmd.SetOutput(out)

	cmd.Flags().String("docker-host", "127.0.0.1", "Docker host/IP. It must be reachable from Dispatch Server.")
	cmd.Flags().Int("gateway-port", 8081, "Port for local API Gateway")
	cmd.Flags().Int("gateway-tls-port", 8444, "TLS port for local API Gateway (only when TLS Enabled in global flags)")
	cmd.Flags().String("api-endpoint", "", "API Endpoint used by event drivers to access Dispatch server.")

	return cmd
}

func initConfig(cmd *cobra.Command, targetConfig *serverConfig) {
	v := viper.New()
	configPath := os.Getenv("DISPATCH_CONFIG")

	if dispatchConfigPath != "" {
		configPath = dispatchConfigPath
	}

	if configPath != "" {
		v.SetConfigFile(dispatchConfigPath)
		if err := v.ReadInConfig(); err != nil {
			log.Fatalf("Unable to read the config file: %s", err)
		}
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		v.BindPFlag(f.Name, f)
		v.BindEnv(f.Name, "DISPATCH_"+strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)))
	})
	err := v.Unmarshal(targetConfig)
	if err != nil {
		log.Fatalf("Unable to create configuration: %s", err)
	}
	if defaultConfig.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func bindLocalFlags(targetStruct interface{}) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		v := viper.New()
		// We use separate viper instance to read service-specific flags, and we must "preload" this instance
		// with values we read from config file, otherwise v.Unmarshal will overwrite them with values from flags
		// even if flags were not used.
		var fromConfig map[string]interface{}
		inrec, _ := json.Marshal(targetStruct)
		json.Unmarshal(inrec, &fromConfig)
		for key, val := range fromConfig {
			v.Set(key, val)
		}
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			v.BindPFlag(f.Name, f)
			v.BindEnv(f.Name, "DISPATCH_"+strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)))
		})
		err := v.Unmarshal(targetStruct)
		if err != nil {
			log.Fatalf("Unable to create configuration: %s", err)
		}
	}
}

func runLocal(config *serverConfig) {
	config.DisableRegistry = true

	store := entityStore(config)
	docker := dockerClient(config)
	functions := functionsClient(config)
	secrets := secretsClient(config)
	images := imagesClient(config)

	secretsService := &service.DBSecretsService{EntityStore: store}
	secretsHandler := initSecrets(config, secretsService)

	imagesHandler, imagesShutdown := initImages(config, store)
	defer imagesShutdown()

	faas := dockerfaas.New(docker)
	functionsDeps := functionsDependencies{
		store:         store,
		faas:          faas,
		dockerclient:  docker,
		imagesClient:  images,
		secretsClient: secrets,
	}
	functionsHandler, functionsShutdown := initFunctions(config, functionsDeps)
	defer functionsShutdown()

	gw, err := local.NewGateway(store, functions)
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
	eventBackend, err := drivers.NewDockerBackend(docker, secrets, config.Local.APIEndpoint)
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
