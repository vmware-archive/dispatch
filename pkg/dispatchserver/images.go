///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"
	"time"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
)

type imageConfig struct {
	PullPeriod time.Duration `mapstructure:"pull-period" json:"pull-period,omitempty"`
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
