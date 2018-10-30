///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager"
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions"
	"github.com/vmware/dispatch/pkg/events"
)

type eventsConfig struct {
	Transport   string `mapstructure:"transport" json:"transport,omitempty"`
	IngressHost string `mapstructure:"ingress-host" json:"ingress-host,omitempty"`
}

type eventsDependencies struct {
	store           entitystore.EntityStore
	transport       events.Transport
	driversBackend  drivers.Backend
	functionsClient client.FunctionsClient
	secretsClient   client.SecretsClient
}

func initEvents(config *serverConfig, deps eventsDependencies) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewEventManagerAPI(swaggerSpec)

	subManager, err := subscriptions.NewManager(deps.transport, deps.functionsClient)
	if err != nil {
		log.Fatalf("Error creating Event Subscription Manager: %v", err)
	}
	// event controller
	eventController := eventmanager.NewEventController(
		subManager,
		deps.driversBackend,
		deps.store,
		eventmanager.EventControllerConfig{
			ResyncPeriod: config.ResyncPeriod,
		},
	)

	eventController.Start()
	// handler
	handlers := &eventmanager.Handlers{
		Store:         deps.store,
		Transport:     deps.transport,
		Watcher:       eventController.Watcher(),
		SecretsClient: deps.secretsClient,
	}

	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		eventController.Shutdown()
		deps.transport.Close()
	}
}
