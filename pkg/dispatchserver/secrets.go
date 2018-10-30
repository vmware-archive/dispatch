///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/secret-store/web"
)

func initSecrets(config *serverConfig, secretsService service.SecretsService) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewSecretStoreAPI(swaggerSpec)

	handlers := web.NewHandlers(secretsService)

	web.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
