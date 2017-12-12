///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/config"
	"github.com/vmware/dispatch/pkg/entity-store"
	iam "github.com/vmware/dispatch/pkg/identity-manager"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/trace"
)

// NewIdentityStore creates a new entitystore for identity manager
func NewIdentityStore() (entitystore.EntityStore, error) {

	kv, err := libkv.NewStore(
		store.BOLTDB,
		[]string{iam.IdentityManagerFlags.DbFile},
		&store.Config{
			Bucket:            "identity",
			ConnectionTimeout: 1 * time.Second,
			PersistConnection: true,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating/opening the entity store:")
	}
	es := entitystore.New(kv)
	return es, nil
}

func init() {
	boltdb.Register()
}

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

	config.Global = config.LoadConfiguration(iam.IdentityManagerFlags.Config)
	config.StaticUsers = config.LoadStaticUsers(iam.IdentityManagerFlags.StaticUsers)

	identityStore, err := NewIdentityStore()
	if err != nil {
		log.Fatalln(err)
	}
	handlers := &iam.Handlers{
		AuthService: iam.NewAuthService(config.Global, identityStore),
	}
	handlers.ConfigureHandlers(api)
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
