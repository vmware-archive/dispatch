///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/secret-store/web"
)

// NewCmdSecrets creates a subcommand to run secrets store
func NewCmdSecrets(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: i18n.T("Run Dispatch Secrets Store"),
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runSecrets(config)
		},
	}
	cmd.SetOutput(out)
	return cmd
}

func runSecrets(config *serverConfig) {
	store := entityStore(config)
	secretsHandler := initSecrets(config, store)

	handler := addMiddleware(secretsHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initSecrets(config *serverConfig, store entitystore.EntityStore) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewSecretStoreAPI(swaggerSpec)

	handlers := web.NewHandlers(&service.DBSecretsService{
		EntityStore: store,
	})

	web.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
