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
	"github.com/vmware/dispatch/pkg/client"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager"
	"github.com/vmware/dispatch/pkg/service-manager/clients"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations"
)

type servicesConfig struct {
	Catalog      string `mapstructure:"catalog" json:"catalog"`
	K8sConfig    string `mapstructure:"kubeconfig" json:"kubeconfig"`
	K8sNamespace string `mapstructure:"namespace" json:"namespace"`
}

// NewCmdServices creates a subcommand to run service manager
func NewCmdServices(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "service-manager",
		Short:  i18n.T("Run Dispatch Service Manager"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Services),
		Run: func(cmd *cobra.Command, args []string) {
			runServices(config)
		},
	}
	cmd.SetOutput(out)

	cmd.Flags().String("kubeconfig", "", "Path to kubernetes config file")
	cmd.Flags().String("namespace", "default", "Kubernetes namespace")
	return cmd
}

func runServices(config *serverConfig) {
	store := entityStore(config)
	secrets := secretsClient(config)

	servicesHandler, servicesShutdown := initServices(config, store, secrets)
	defer servicesShutdown()

	handler := addMiddleware(servicesHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initServices(config *serverConfig, store entitystore.EntityStore, secretsClient client.SecretsClient) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewServiceManagerAPI(swaggerSpec)

	k8sClient, err := clients.NewK8sBrokerClient(
		clients.K8sBrokerConfigOpts{
			K8sConfig:        config.Services.K8sConfig,
			CatalogNamespace: config.Services.K8sNamespace,
			SecretsClient:    secretsClient,
		},
	)
	if err != nil {
		log.Fatalf("Error creating k8sClient: %v", err)
	}

	controller := servicemanager.NewController(
		&servicemanager.ControllerConfig{
			ResyncPeriod: config.ResyncPeriod,
		},
		store,
		k8sClient,
	)
	controller.Start()

	// handler
	handlers := &servicemanager.Handlers{
		Store:   store,
		Watcher: controller.Watcher(),
	}

	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		controller.Shutdown()
	}
}
