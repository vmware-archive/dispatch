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
	"github.com/vmware/dispatch/pkg/identity-manager"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
)

type identityConfig struct {
	CookieName          string `mapstructure:"cookie-name" json:"cookie-name,omitempty"`
	SkipAuth            bool   `mapstructure:"skip-auth" json:"skip-auth,omitempty"`
	BootstrapConfigPath string `mapstructure:"bootstrap-config-path" json:"bootstrap-config-path,omitempty"`
	OAuth2ProxyAuthURL  string `mapstructure:"oauth2-proxy-auth-url" json:"oauth2-proxy-auth-url,omitempty"`
}

// NewCmdIdentity creates a subcommand to run identity manager
func NewCmdIdentity(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "identity-manager",
		Short:  i18n.T("Run Dispatch Identity Manager"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Identity),
		Run: func(cmd *cobra.Command, args []string) {
			runIdentity(config)
		},
	}
	cmd.SetOutput(out)

	cmd.Flags().String("cookie-name", "_oauth2_proxy", "The cookie name used to identify users")
	cmd.Flags().Bool("skip-auth", false, "Skips authorization, not to be used in production env")
	cmd.Flags().String("bootstrap-config-path", "/bootstrap", "The path that contains the bootstrap keys")
	cmd.Flags().String("oauth2-proxy-auth-url", "http://localhost:4180/v1/iam/oauth2/auth", "The localhost url for oauth2proxy service's auth endpoint")

	return cmd
}

func runIdentity(config *serverConfig) {
	store := entityStore(config)
	identityHandler, shutdown := initIdentity(config, store)
	defer shutdown()

	handler := addMiddleware(identityHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initIdentity(config *serverConfig, store entitystore.EntityStore) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewIdentityManagerAPI(swaggerSpec)

	// Setup the policy enforcer
	enforcer := identitymanager.SetupEnforcer(store)

	// Create the identity controller
	controller := identitymanager.NewIdentityController(store, enforcer, config.ResyncPeriod)
	controller.Start()

	handlers := identitymanager.NewHandlers(controller.Watcher(), store, enforcer)
	handlers.ConfigureHandlers(api)
	handlers.CookieName = config.Identity.CookieName
	handlers.BootstrapConfigPath = config.Identity.BootstrapConfigPath
	handlers.OAuth2ProxyAuthURL = config.Identity.OAuth2ProxyAuthURL
	handlers.SkipAuth = config.Identity.SkipAuth

	return api.Serve(nil), func() {
		controller.Shutdown()
	}
}
