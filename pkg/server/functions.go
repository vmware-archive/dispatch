///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/gen/restapi"
	"github.com/vmware/dispatch/pkg/functions/gen/restapi/operations"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

func initFunctions(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionsAPI(swaggerSpec)
	handlers := functions.NewHandlers(
		config.K8sConfig, config.Namespace, config.ImageRegistry, config.SourceRoot)
	functions.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
