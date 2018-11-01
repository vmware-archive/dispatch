///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	dockerclient "github.com/docker/docker/client"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/injectors"
	"github.com/vmware/dispatch/pkg/functions/runner"
	"github.com/vmware/dispatch/pkg/functions/validator"
	"github.com/vmware/dispatch/pkg/utils"
)

type functionsConfig struct {
	ImagePullSecret     string                       `mapstructure:"image-pull-secret" json:"image-pull-secret,omitempty"`
	FuncDefaultLimits   *functions.FunctionResources `mapstructure:"func-default-limits" json:"func-default-limits,omitempty"`
	FuncDefaultRequests *functions.FunctionResources `mapstructure:"func-default-requests" json:"func-default-requests,omitempty"`
	FileImageManager    string                       `mapstructure:"file-image-manager" json:"file-image-manager,omitempty"`
}

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

type functionsDependencies struct {
	store         entitystore.EntityStore
	faas          functions.FaaSDriver
	dockerclient  dockerclient.CommonAPIClient
	imagesClient  functionmanager.ImageGetter
	secretsClient client.SecretsClient
}

func initFunctions(config *serverConfig, deps functionsDependencies) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionManagerAPI(swaggerSpec)

	c := &functionmanager.ControllerConfig{
		ResyncPeriod: config.ResyncPeriod,
	}

	r := runner.New(&runner.Config{
		Faas:           deps.faas,
		Validator:      validator.New(),
		SecretInjector: injectors.NewSecretInjector(deps.secretsClient),
	})

	imageBuilder := functions.NewDockerImageBuilder(config.ImageRegistry, config.RegistryAuth, deps.dockerclient)
	if config.DisableRegistry {
		imageBuilder.PushImages = false
		imageBuilder.PullImages = false
	}

	controller := functionmanager.NewController(c, deps.store, deps.faas, r, deps.imagesClient, imageBuilder)
	controller.Start()

	handlers := functionmanager.NewHandlers(controller.Watcher(), deps.store)
	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		controller.Shutdown()
		utils.Close(deps.faas)
	}
}
