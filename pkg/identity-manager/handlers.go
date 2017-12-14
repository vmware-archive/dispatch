///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"fmt"
	"net/http"
	"net/url"

	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
)

// IdentityManagerFlags are configuration flags for the identity manager
var IdentityManagerFlags = struct {
	CookieName string `long:"cookie-name" description:"The cookie name used to identify users" default:"_oauth2_proxy"`
}{}

type Handlers struct{}

// ConfigureHandlers registers the identity manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {

	a, ok := api.(*operations.IdentityManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuth = func(token string) (interface{}, error) {

		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.HomeHandler = operations.HomeHandlerFunc(h.home)
	a.RootHandler = operations.RootHandlerFunc(h.root)
	a.RedirectHandler = operations.RedirectHandlerFunc(h.redirect)
}

func (h *Handlers) home(params operations.HomeParams, principal interface{}) middleware.Responder {

	message := "Home Page, You have already logged in"
	return operations.NewHomeOK().WithPayload(
		&models.Message{Message: swag.String(message)})
}

func (h *Handlers) root(params operations.RootParams) middleware.Responder {
	message := "Default Root Page, no authentication required"
	return operations.NewRootOK().WithPayload(
		&models.Message{Message: swag.String(message)})
}

func (h *Handlers) redirect(params operations.RedirectParams, principal interface{}) middleware.Responder {

	redirect := *params.Redirect
	cookie, err := params.HTTPRequest.Cookie(IdentityManagerFlags.CookieName)
	if err != nil {
		return operations.NewRedirectDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{Code: http.StatusInternalServerError,
				Message: swag.String("No Such Cookie")})
	}

	values := url.Values{
		"cookie": {cookie.String()},
	}
	location := fmt.Sprintf("%s?%s", redirect, values.Encode())
	return operations.NewRedirectFound().WithLocation(location)
}
