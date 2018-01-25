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

	applicationmanager "github.com/vmware/dispatch/pkg/application-manager"
	"github.com/vmware/dispatch/pkg/application-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/application-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/entity-store"
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
		swag.CommandLineOptionsGroup{
			ShortDescription: "Application Manager Flags",
			LongDescription:  "",
			Options:          &applicationmanager.ApplicationManagerFlags,
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

	app := operations.NewApplicationManagerAPI(swaggerSpec)
	server := restapi.NewServer(app)

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Application Manager"
	parser.LongDescription = "This is the API server for the Dispatch Application Manager service.\n"

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

	// entity store
	es, err := entitystore.NewFromBackend(
		entitystore.BackendConfig{
			Backend:  applicationmanager.ApplicationManagerFlags.DbBackend,
			Address:  applicationmanager.ApplicationManagerFlags.DbFile,
			Bucket:   applicationmanager.ApplicationManagerFlags.DbDatabase,
			Username: applicationmanager.ApplicationManagerFlags.DbUser,
			Password: applicationmanager.ApplicationManagerFlags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}
	// handlers
	handlers := applicationmanager.NewHandlers(nil, es)
	handlers.ConfigureHandlers(app)

	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	handler := alice.New(
		middleware.NewHealthCheckMW("", healthChecker),
	).Then(app.Serve(nil))

	server.SetHandler(handler)

	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
