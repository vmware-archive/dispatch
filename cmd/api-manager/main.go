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
	log "github.com/sirupsen/logrus"

	apimanager "github.com/vmware/dispatch/pkg/api-manager"
	"github.com/vmware/dispatch/pkg/api-manager/gateway/kong"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations"
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
			ShortDescription: "API Manager Flags",
			LongDescription:  "",
			Options:          &apimanager.APIManagerFlags,
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

	api := operations.NewAPIManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "API Manager"
	parser.LongDescription = "This is the API server for the Dispatch API Manager service.\n"

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
			Backend:  apimanager.APIManagerFlags.DbBackend,
			Address:  apimanager.APIManagerFlags.DbFile,
			Bucket:   apimanager.APIManagerFlags.DbDatabase,
			Username: apimanager.APIManagerFlags.DbUser,
			Password: apimanager.APIManagerFlags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}

	// api gateway
	gateway, err := kong.NewClient(&kong.Config{
		Host:     apimanager.APIManagerFlags.GatewayHost,
		Upstream: apimanager.APIManagerFlags.FunctionManager,
	})
	if err != nil {
		log.Fatalf("Error creating an api gateway client: %v", err)
	}

	// controller
	config := &apimanager.ControllerConfig{
		ResyncPeriod:   time.Duration(apimanager.APIManagerFlags.ResyncPeriod) * time.Second,
		OrganizationID: apimanager.APIManagerFlags.OrgID,
	}
	controller := apimanager.NewController(config, es, gateway)
	defer controller.Shutdown()
	controller.Start()

	// handlers
	handlers := apimanager.NewHandlers(controller.Watcher(), es)
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
