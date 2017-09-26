///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package main

import (
	"log"
	"os"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/swag"
	flags "github.com/jessevdk/go-flags"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
	boltdb.Register()
}

func configureFlags() []swag.CommandLineOptionsGroup {
	return []swag.CommandLineOptionsGroup{
		swag.CommandLineOptionsGroup{
			ShortDescription: "Function manager Flags",
			LongDescription:  "",
			Options:          &functionmanager.FunctionManagerFlags,
		},
	}
}

func main() {

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionManagerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Function Manager"
	parser.LongDescription = "This is the API server for the serverless function manager service.\n"

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

	kv, err := libkv.NewStore(
		store.BOLTDB,
		[]string{functionmanager.FunctionManagerFlags.DbFile},
		&store.Config{
			Bucket:            "function",
			ConnectionTimeout: 1 * time.Second,
			PersistConnection: true,
		},
	)
	if err != nil {
		log.Fatalf("Error creating/opening the entity store: %v", err)
	}
	es := entitystore.New(kv)

	functionmanager.ConfigureHandlers(api, es)

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
