///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"os"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag"
	"github.com/jessevdk/go-flags"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager"
	iam "github.com/vmware/dispatch/pkg/identity-manager"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/trace"
)

var debugFlags = struct {
	DebugEnabled   bool `long:"debug" description:"Enable debugging messages"`
	TracingEnabled bool `long:"trace" description:"Enable tracing messages (enables debugging)"`
}{}

func configureFlags() []swag.CommandLineOptionsGroup {
	return []swag.CommandLineOptionsGroup{
		swag.CommandLineOptionsGroup{
			ShortDescription: "Identity Manager Flags",
			LongDescription:  "",
			Options:          &iam.IdentityManagerFlags,
		},
		swag.CommandLineOptionsGroup{
			ShortDescription: "Debug options",
			LongDescription:  "",
			Options:          &debugFlags,
		},
	}
}

func main() {

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewIdentityManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Identity Manager"
	parser.LongDescription = "This is the API server for the Dispatch Identity Manager service.\n"

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
			Backend:  identitymanager.IdentityManagerFlags.DbBackend,
			Address:  identitymanager.IdentityManagerFlags.DbFile,
			Bucket:   identitymanager.IdentityManagerFlags.DbDatabase,
			Username: identitymanager.IdentityManagerFlags.DbUser,
			Password: identitymanager.IdentityManagerFlags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}

	// Setup the policy enforcer
	enforcer := identitymanager.SetupEnforcer(es)

	// Create the identity controller
	controller := identitymanager.NewIdentityController(es, enforcer)
	defer controller.Shutdown()
	controller.Start()

	handlers := identitymanager.NewHandlers(controller.Watcher(), es, enforcer)
	handlers.ConfigureHandlers(api)

	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	handler := alice.New(
		middleware.NewHealthCheckMW("", healthChecker),
	).Then(api.Serve(nil))

	server.SetHandler(handler)

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
