///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"net/http"
	"time"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
)

type imageConfig struct {
	PullPeriod time.Duration `mapstructure:"pull-period" json:"pull-period,omitempty"`
}

// NewCmdImages creates a subcommand to run image manager
func NewCmdImages(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "image-manager",
		Short:  i18n.T("Run Dispatch Image Manager"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Image),
		Run: func(cmd *cobra.Command, args []string) {
			runImages(config)
		},
	}
	cmd.SetOutput(out)

	cmd.Flags().Duration("pull-period", time.Duration(3*time.Hour), "How often images should be pulled to verify that they exist on the registry")
	return cmd
}

func runImages(config *serverConfig) {
	store := entityStore(config)
	imagesHandler, shutdown := initImages(config, store)
	defer shutdown()

	handler := addMiddleware(imagesHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initImages(config *serverConfig, store entitystore.EntityStore) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewImageManagerAPI(swaggerSpec)

	c := &imagemanager.ControllerConfig{
		ResyncPeriod: config.ResyncPeriod,
	}

	registryAuth := config.RegistryAuth
	if registryAuth == "" {
		registryAuth = emptyRegistryAuth
	}

	ib, err := imagemanager.NewImageBuilder(store, config.ImageRegistry, registryAuth, config.Image.PullPeriod)
	if err != nil {
		log.Fatalln(err)
	}
	if config.DisableRegistry {
		ib.PushImages = false
	}
	bib, err := imagemanager.NewBaseImageBuilder(store, config.Image.PullPeriod)
	if err != nil {
		log.Fatalln(err)
	}

	controller := imagemanager.NewController(c, store, bib, ib)
	controller.Start()

	handlers := imagemanager.NewHandlers(ib, bib, controller.Watcher(), store)
	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		controller.Shutdown()
	}
}
