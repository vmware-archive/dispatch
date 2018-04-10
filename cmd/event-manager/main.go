///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"os"

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
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/transport"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/trace"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
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
			Options:          &eventmanager.Flags,
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

	config.Global = config.LoadConfiguration(eventmanager.Flags.Config)

	store, err := entitystore.NewFromBackend(
		entitystore.BackendConfig{
			Backend:  eventmanager.Flags.DbBackend,
			Address:  eventmanager.Flags.DbFile,
			Bucket:   eventmanager.Flags.DbDatabase,
			Username: eventmanager.Flags.DbUser,
			Password: eventmanager.Flags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}

	var eventTransport events.Transport

	switch eventmanager.Flags.Transport {
	// TODO: make transport types constants/iota
	case "kafka":
		eventTransport, err = transport.NewKafka(eventmanager.Flags.KafkaBrokers)
		if err != nil {
			log.Fatalf("Error creating Kafka event transport: %+v", err)
		}
	case "rabbitmq":
		eventTransport, err = transport.NewRabbitMQ(
			eventmanager.Flags.RabbitMQURL,
			eventmanager.Flags.OrgID,
		)
		if err != nil {
			log.Fatalf("Error creating RabbitMQ event transport: %+v", err)
		}
	default:
		log.Fatalf("Transport %s is not supported. pick one of [kafka,rabbitmq]", eventmanager.Flags.Transport)
	}
	defer eventTransport.Close()

	fnClient := client.NewFunctionsClient(eventmanager.Flags.FunctionManager, client.AuthWithToken("cookie"))

	subManager, err := subscriptions.NewManager(eventTransport, fnClient)
	if err != nil {
		log.Fatalf("Error creating SubscriptionManager: %v", err)
	}

	k8sBackend, err := drivers.NewK8sBackend(
		drivers.ConfigOpts{
			DriverImage:     eventmanager.Flags.EventDriverImage,
			SidecarImage:    eventmanager.Flags.EventSidecarImage,
			TransportType:   eventmanager.Flags.Transport,
			KafkaBrokers:    eventmanager.Flags.KafkaBrokers,
			RabbitMQURL:     eventmanager.Flags.RabbitMQURL,
			TracerURL:       eventmanager.Flags.TracerURL,
			K8sConfig:       eventmanager.Flags.K8sConfig,
			DriverNamespace: eventmanager.Flags.K8sNamespace,
			SecretStoreURL:  eventmanager.Flags.SecretStore,
			OrgID:           eventmanager.Flags.OrgID,
		},
	)
	if err != nil {
		log.Fatalf("Error creating k8sBackend: %v", err)
	}
	// event controller
	eventController := eventmanager.NewEventController(
		subManager,
		k8sBackend,
		store,
		eventmanager.EventControllerConfig{OrganizationID: eventmanager.Flags.OrgID},
	)

	defer eventController.Shutdown()
	eventController.Start()

	// handler
	handlers := &eventmanager.Handlers{
		Store:     store,
		Transport: eventTransport,
		Watcher:   eventController.Watcher(),
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
