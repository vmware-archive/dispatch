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
	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/swag"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	imagemanager "github.com/vmware/dispatch/pkg/image-manager"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/image-manager/gen/restapi/operations"
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

	kv, err := libkv.NewStore(
		store.BOLTDB,
		[]string{imagemanager.ImageManagerFlags.DbFile},
		&store.Config{
			Bucket:            "image",
			ConnectionTimeout: 1 * time.Second,
			PersistConnection: true,
		},
	)
	if err != nil {
		log.Fatalf("Error creating/opening the entity store: %v", err)
	}
	es := entitystore.New(kv)

	ib, err := imagemanager.NewImageBuilder(es)
	if err != nil {
		log.Fatalln(err)
	}
	bib, err := imagemanager.NewBaseImageBuilder(es)
	if err != nil {
		log.Fatalln(err)
	}

	handlers := imagemanager.NewHandlers(ib, bib, es)

	go ib.Run()
	go bib.Run()

	handlers.ConfigureHandlers(api)

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
