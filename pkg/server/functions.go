///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	apiclient "github.com/go-openapi/runtime/client"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/functions"
	fconfig "github.com/vmware/dispatch/pkg/functions/config"
	"github.com/vmware/dispatch/pkg/functions/gen/restapi"
	"github.com/vmware/dispatch/pkg/functions/gen/restapi/operations"
)

const defaultBuildImage = "dispatchframework/dispatch-knative-builder:0.0.2"

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

func initFunctions(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionsAPI(swaggerSpec)
	// TODO (bjung): why is the client bound to an org ID?
	// TODO: address dummy auth
	auth := apiclient.APIKeyAuth("cookie", "header", "UNSET")
	imagesClient := client.NewImagesClient(fmt.Sprintf("localhost:%d", config.Port), auth, config.Namespace)

	var storageConfig *fconfig.StorageConfig
	switch config.Storage {
	case string(fconfig.Minio):
		storageConfig = &fconfig.StorageConfig{
			Storage: fconfig.Minio,
			Minio: &fconfig.StorageMinioConfig{
				MinioAddress: config.MinioAddress,
				Username:     config.MinioUsername,
				Password:     config.MinioPassword,
				Location:     fconfig.DefaultMinioLocation,
			},
		}
	case string(fconfig.File):
		storageConfig = &fconfig.StorageConfig{
			Storage: fconfig.File,
			File: &fconfig.StorageFileConfig{
				SourceRootPath: config.FileSourceRoot,
			},
		}
	default:
		log.Fatalf("incompatible storage config %s", config.Storage)
	}

	handlers := functions.NewHandlers(
		config.K8sConfig, config.Namespace, config.ImageRegistry, config.IngressGatewayIP, config.BuildImage, storageConfig, imagesClient)
	functions.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
