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
	"github.com/vmware/dispatch/pkg/event-manager"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions"
	"github.com/vmware/dispatch/pkg/events/transport"
)

// NewCmdEvents creates a subcommand to run event manager
func NewCmdEvents(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: i18n.T("Run Dispatch Event Manager"),
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runEvents(config)
		},
	}
	cmd.SetOutput(out)
	return cmd
}

func runEvents(config *serverConfig) {
	store := entityStore(config)
	functions := functionsClient(config)
	secrets := secretsClient(config)
	eventsHandler, shutdown := initEvents(config, store, functions, secrets)
	defer shutdown()

	handler := addMiddleware(eventsHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initEvents(config *serverConfig, store entitystore.EntityStore, fnClient client.FunctionsClient, secretsClient client.SecretsClient) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewEventManagerAPI(swaggerSpec)

	eventTransport := transport.NewInMemory()

	subManager, err := subscriptions.NewManager(eventTransport, fnClient)
	if err != nil {
		log.Fatalf("Error creating SubscriptionManager: %v", err)
	}
	// event controller
	eventController := eventmanager.NewEventController(
		subManager,
		// TODO: add backend for event drivers in docker
		nil,
		store,
		eventmanager.EventControllerConfig{},
	)

	eventController.Start()
	// handler
	handlers := &eventmanager.Handlers{
		Store:         store,
		Transport:     eventTransport,
		Watcher:       eventController.Watcher(),
		SecretsClient: secretsClient,
	}

	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		eventController.Shutdown()
		eventTransport.Close()
	}
}
