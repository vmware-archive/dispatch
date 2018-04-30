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

	"github.com/vmware/dispatch/pkg/config"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/image-manager"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
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
		swag.CommandLineOptionsGroup{
			ShortDescription: "Image Manager Flags",
			LongDescription:  "",
			Options:          &imagemanager.ImageManagerFlags,
		},
		swag.CommandLineOptionsGroup{
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

	api := operations.NewImageManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Image Manager"
	parser.LongDescription = "This is the API server for the Dispatch Image Manager service.\n"

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

	config.Global = config.LoadConfiguration(imagemanager.ImageManagerFlags.Config)

	es, err := entitystore.NewFromBackend(
		entitystore.BackendConfig{
			Backend:  imagemanager.ImageManagerFlags.DbBackend,
			Address:  imagemanager.ImageManagerFlags.DbFile,
			Bucket:   imagemanager.ImageManagerFlags.DbDatabase,
			Username: imagemanager.ImageManagerFlags.DbUser,
			Password: imagemanager.ImageManagerFlags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}

	c := &imagemanager.ControllerConfig{
		ResyncPeriod:   time.Duration(imagemanager.ImageManagerFlags.ResyncPeriod) * time.Second,
		OrganizationID: imagemanager.ImageManagerFlags.OrgID,
	}

	registryAuth := config.Global.Registry.RegistryAuth
	if config.Global.Registry.RegistryAuth == "" {
		registryAuth = config.EmptyRegistryAuth
	}

	ib, err := imagemanager.NewImageBuilder(es, config.Global.Registry.RegistryURI, registryAuth)
	if err != nil {
		log.Fatalln(err)
	}
	bib, err := imagemanager.NewBaseImageBuilder(es)
	if err != nil {
		log.Fatalln(err)
	}

	controller := imagemanager.NewController(c, es, bib, ib)
	defer controller.Shutdown()
	controller.Start()

	handlers := imagemanager.NewHandlers(ib, bib, controller.Watcher(), es)
	handlers.ConfigureHandlers(api)

	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	tracer, tracingCloser, err := utils.CreateTracer("ImageManager", imagemanager.ImageManagerFlags.Tracer)
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
