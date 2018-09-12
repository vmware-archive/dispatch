///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/http"
)

func runDispatch(config *serverConfig) {

	functionsHandler := initFunctions(config)
	secretsHandler := initSecrets(config)

	dispatchHandler := &http.AllInOneRouter{
		FunctionsHandler: functionsHandler,
		SecretsHandler:   secretsHandler,
	}
	handler := addMiddleware(dispatchHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}
