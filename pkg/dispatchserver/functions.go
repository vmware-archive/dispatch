///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"net/http"

	dockerclient "github.com/docker/docker/client"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/docker"
	"github.com/vmware/dispatch/pkg/functions/injectors"
	"github.com/vmware/dispatch/pkg/functions/runner"
	"github.com/vmware/dispatch/pkg/functions/validator"
	"github.com/vmware/dispatch/pkg/utils"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

// NewCmdFunctions creates a subcommand to create functions manager
func NewCmdFunctions(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "functions",
		Short: i18n.T("Run Dispatch Functions Manager"),
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runFunctions(config)
		},
	}
	cmd.SetOutput(out)
	return cmd
}

func runFunctions(config *serverConfig) {
	store := entityStore(config)
	docker := dockerClient(config)
	secrets := secretsClient(config)
	services := servicesClient(config)
	images := imagesClient(config)

	fnHandler, shutdown := initFunctions(config, store, docker, images, secrets, services)
	defer shutdown()

	handler := addMiddleware(fnHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initFunctions(
	config *serverConfig, store entitystore.EntityStore, dockerclient dockerclient.CommonAPIClient, imagesClient client.ImagesClient,
	secretsClient client.SecretsClient, servicesClient client.ServicesClient) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionManagerAPI(swaggerSpec)

	faas := docker.New(dockerclient)

	c := &functionmanager.ControllerConfig{
		ResyncPeriod: config.ResyncPeriod,
	}

	r := runner.New(&runner.Config{
		Faas:            faas,
		Validator:       validator.New(),
		SecretInjector:  injectors.NewSecretInjector(secretsClient),
		ServiceInjector: injectors.NewServiceInjector(secretsClient, servicesClient),
	})

	imageBuilder := functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, dockerclient)
	if !config.PushImages {
		imageBuilder.PushImages = false
		imageBuilder.PullImages = false
	}

	controller := functionmanager.NewController(c, store, faas, r, imagesClient, imageBuilder)
	controller.Start()

	handlers := functionmanager.NewHandlers(controller.Watcher(), store)
	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		controller.Shutdown()
		utils.Close(faas)
	}
}
