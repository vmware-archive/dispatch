///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
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
