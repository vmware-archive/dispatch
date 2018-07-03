///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/swag"
	"github.com/jessevdk/go-flags"
	"github.com/justinas/alice"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/secret-store/web"
	"github.com/vmware/dispatch/pkg/utils"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

var debugFlags = struct {
	DebugEnabled bool `long:"debug" description:"Enable debugging messages"`
}{}

func configureFlags() []swag.CommandLineOptionsGroup {
	return []swag.CommandLineOptionsGroup{
		swag.CommandLineOptionsGroup{
			ShortDescription: "Secret Store Flags",
			LongDescription:  "",
			Options:          &web.SecretStoreFlags,
		},
		swag.CommandLineOptionsGroup{
			ShortDescription: "Debug options",
			LongDescription:  "",
			Options:          &debugFlags,
		},
	}
}

func main() {

	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewSecretStoreAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Secret Store"
	parser.LongDescription = "An API for managing secrets for Dispatch."

	optsGroups := configureFlags()
	for _, optsGroup := range optsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if debugFlags.DebugEnabled {
		log.SetLevel(log.DebugLevel)
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

	entityStore, err := entitystore.NewFromBackend(
		entitystore.BackendConfig{
			Backend:  web.SecretStoreFlags.DbBackend,
			Address:  web.SecretStoreFlags.DbFile,
			Bucket:   web.SecretStoreFlags.DbDatabase,
			Username: web.SecretStoreFlags.DbUser,
			Password: web.SecretStoreFlags.DbPassword,
		})
	if err != nil {
		log.Fatalln(err)
	}

	var config *rest.Config
	if web.SecretStoreFlags.K8sConfig == "" {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", web.SecretStoreFlags.K8sConfig)
	}
	if err != nil {
		log.Fatalf("Error getting kubernetes config: %+v", err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %+v", err)
	}

	handlers := web.NewHandlers(&service.K8sSecretsService{
		EntityStore: entityStore,
		SecretsAPI:  clientset.CoreV1().Secrets(web.SecretStoreFlags.K8sNamespace),
	})

	web.ConfigureHandlers(api, handlers)

	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	tracer, tracingCloser, err := utils.CreateTracer("SecretStore", web.SecretStoreFlags.Tracer)
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

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
