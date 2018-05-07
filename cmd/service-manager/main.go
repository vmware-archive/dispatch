///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"os"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/swag"
	"github.com/jessevdk/go-flags"
	"github.com/justinas/alice"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/utils"

	"github.com/vmware/dispatch/pkg/config"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/service-manager"
	"github.com/vmware/dispatch/pkg/service-manager/clients"
	servicemanagerflags "github.com/vmware/dispatch/pkg/service-manager/flags"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations"
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
			ShortDescription: "Service manager Flags",
			LongDescription:  "",
			Options:          &servicemanagerflags.ServiceManagerFlags,
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

	api := operations.NewServiceManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Service Manager"
	parser.LongDescription = "This is the API server for the Dispatch Service Manager service.\n"

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

	config.Global = config.LoadConfiguration(servicemanagerflags.ServiceManagerFlags.Config)

	store, err := entitystore.NewFromBackend(
		entitystore.BackendConfig{
			Backend:  servicemanagerflags.ServiceManagerFlags.DbBackend,
			Address:  servicemanagerflags.ServiceManagerFlags.DbFile,
			Bucket:   servicemanagerflags.ServiceManagerFlags.DbDatabase,
			Username: servicemanagerflags.ServiceManagerFlags.DbUser,
			Password: servicemanagerflags.ServiceManagerFlags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}

	k8sClient, err := clients.NewK8sBrokerClient(
		clients.K8sBrokerConfigOpts{
			K8sConfig:        servicemanagerflags.ServiceManagerFlags.K8sConfig,
			CatalogNamespace: config.Global.Service.K8sServiceCatalog.CatalogNamespace,
			SecretStoreURL:   servicemanagerflags.ServiceManagerFlags.SecretStore,
			OrgID:            servicemanagerflags.ServiceManagerFlags.OrgID,
		},
	)
	if err != nil {
		log.Fatalf("Error creating k8sClient: %v", err)
	}
	// service controller
	serviceController := servicemanager.NewController(
		&servicemanager.ControllerConfig{
			OrganizationID: servicemanagerflags.ServiceManagerFlags.OrgID,
			ResyncPeriod:   time.Second * time.Duration(servicemanagerflags.ServiceManagerFlags.ResyncPeriod),
		},
		store,
		k8sClient,
	)

	defer serviceController.Shutdown()
	serviceController.Start()

	// handler
	handlers := &servicemanager.Handlers{
		Store:   store,
		Watcher: serviceController.Watcher(),
	}

	handlers.ConfigureHandlers(api)

	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	tracer, tracingCloser, err := utils.CreateTracer("ServiceManager", servicemanagerflags.ServiceManagerFlags.Tracer)
	if err != nil {
		log.Fatalf("Error creating a tracer: %+v", err)
	}
	defer tracingCloser.Close()
	opentracing.SetGlobalTracer(tracer)

	handler := alice.New(
		middleware.NewHealthCheckMW("", healthChecker),
		middleware.NewTracingMW(tracer),
	).Then(api.Serve(nil))

	server.SetHandler(handler)

	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
