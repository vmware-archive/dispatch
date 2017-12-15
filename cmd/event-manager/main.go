///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"os"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/swag"
	"github.com/jessevdk/go-flags"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/config"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/events/rabbitmq"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/trace"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
	boltdb.Register()
}

var debugFlags = struct {
	DebugEnabled   bool `long:"debug" description:"Enable debugging messages"`
	TracingEnabled bool `long:"trace" description:"Enable tracing messages (enables debugging)"`
}{}

func configureFlags() []swag.CommandLineOptionsGroup {
	return []swag.CommandLineOptionsGroup{
		{
			ShortDescription: "Event manager Flags",
			LongDescription:  "",
			Options:          &eventmanager.EventManagerFlags,
		},
		{
			ShortDescription: "Debug options",
			LongDescription:  "",
			Options:          &debugFlags,
		},
	}
}

func main() {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewEventManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Event Manager"
	parser.LongDescription = "This is the API server for the Dispatch Event Manager service.\n"

	optsGroups := configureFlags()
	for _, optsGroup := range optsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	if debugFlags.DebugEnabled {
		log.SetLevel(log.DebugLevel)
	}
	if debugFlags.TracingEnabled {
		log.SetLevel(log.DebugLevel)
		trace.Enable()
	}

	config.Global = config.LoadConfiguration(eventmanager.EventManagerFlags.Config)

	kv, err := libkv.NewStore(
		store.BOLTDB,
		[]string{eventmanager.EventManagerFlags.DbFile},
		&store.Config{
			Bucket:            "function",
			ConnectionTimeout: 1 * time.Second,
			PersistConnection: true,
		},
	)
	if err != nil {
		log.Fatalf("Error creating/opening the entity store: %v", err)
	}

	store := entitystore.New(kv)

	// TODO: add more parameters to be customizable via flags
	queue, err := rabbitmq.New(
		eventmanager.EventManagerFlags.AMQPURL,
		"dispatch",
	)
	if err != nil {
		log.Fatalf("Error creating RabbitMQ connection: %+v", err)
	}
	defer queue.Close()

	fnClient := client.NewFunctionsClient(eventmanager.EventManagerFlags.FunctionManager, client.AuthWithToken("cookie"))

	subManager, err := eventmanager.NewSubscriptionManager(queue, fnClient)
	if err != nil {
		log.Fatalf("Error creating SubscriptionManager: %v", err)
	}

	k8sBackend, err := eventmanager.NewK8sBackend()
	if err != nil {
		log.Fatalf("Error creating k8sBackend: %v", err)
	}
	// event controller
	eventController := eventmanager.NewEventController(
		subManager,
		k8sBackend,
		store,
		eventmanager.EventControllerConfig{OrganizationID: eventmanager.EventManagerFlags.OrgID},
	)

	defer eventController.Shutdown()
	eventController.Start()

	// handler
	handlers := &eventmanager.Handlers{
		Store:   store,
		EQ:      queue,
		Watcher: eventController.Watcher(),
	}

	handlers.ConfigureHandlers(api)

	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	handler := alice.New(
		middleware.NewHealthCheckMW("", healthChecker),
	).Then(api.Serve(nil))

	server.SetHandler(handler)

	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
