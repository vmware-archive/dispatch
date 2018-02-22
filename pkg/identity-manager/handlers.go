///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/casbin/casbin"
	casbinFileAdapter "github.com/casbin/casbin/file-adapter"
	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
)

// IdentityManagerFlags are configuration flags for the identity manager
var IdentityManagerFlags = struct {
	PolicyFile string `long:"policy-file" description:"Path to policy file" default:"./policy.csv"`
	CookieName string `long:"cookie-name" description:"The cookie name used to identify users" default:"_oauth2_proxy"`
	SkipAuth   bool   `long:"skip-auth" description:"Skips authorization, not to be used in production env"`
}{}

const (
	// Policy Model - Use an ACL model that matches request attributes
	casbinPolicyModel = `
[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = keyMatch(r.sub, p.sub) && keyMatch(r.obj, p.obj) && keyMatch(r.act, p.act)
`
)

const (
	HTTP_HEADER_ORIG_URL    = "X-Original-URL"
	HTTP_HEADER_ORIG_METHOD = "X-Original-Method"
	HTTP_HEADER_FWD_EMAIL   = "X-Forwarded-Email"
)

// Identity manager action constants
const (
	ActionGet    Action = "get"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

var (
	model = casbin.NewModel(casbinPolicyModel)
)

// Action defines the type for an action
type Action string

// Handlers defines the interface for the identity manager handlers
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

	a.RootHandler = operations.RootHandlerFunc(h.root)
	a.HomeHandler = operations.HomeHandlerFunc(h.home)
	a.AuthHandler = operations.AuthHandlerFunc(h.auth)
	a.RedirectHandler = operations.RedirectHandlerFunc(h.redirect)
}

func (h *Handlers) root(params operations.RootParams) middleware.Responder {
	message := "Default Root Page"
	return operations.NewRootOK().WithPayload(
		&models.Message{Message: swag.String(message)})
}

func (h *Handlers) home(params operations.HomeParams, principal interface{}) middleware.Responder {

	message := "Home Page, You have already logged in"
	return operations.NewHomeOK().WithPayload(
		&models.Message{Message: swag.String(message)})
}

func (h *Handlers) auth(params operations.AuthParams, principal interface{}) middleware.Responder {
	// For development use cases, not recommended in production env.
	if IdentityManagerFlags.SkipAuth {
		log.Warn("Skipping authorization. This is not recommended in production environments.")
		return operations.NewAuthAccepted()
	}
	// At this point, the user is authenticated, let's do a policy check.
	log.Debugf("Loading policies from file %s", IdentityManagerFlags.PolicyFile)
	adapter := casbinFileAdapter.NewAdapter(IdentityManagerFlags.PolicyFile)
	enforcer := casbin.NewEnforcer(model, adapter)
	attrs, err := getRequestAttributes(params.HTTPRequest)
	if err != nil {
		log.Debugf("Unable to parse request attributes: %s", err)
		return operations.NewAuthForbidden()
	}
	log.Debugf("Enforcing Policy: %s, %s, %s\n", attrs.userEmail, attrs.resource, attrs.action)

	// Note: Non-Resource requests are currently not authz enforced.
	if !attrs.isResourceRequest {
		return operations.NewAuthAccepted()
	}

	if enforcer.Enforce(attrs.userEmail, attrs.resource, string(attrs.action)) == true {
		return operations.NewAuthAccepted()
	}

	// deny the request, show an error
	return operations.NewAuthForbidden()
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

func getRequestAttributes(request *http.Request) (*attributesRecord, error) {
	log.Debugf("Headers: %s\n", request.Header)

	// Get User Info
	userEmail := request.Header.Get(HTTP_HEADER_FWD_EMAIL)
	if userEmail == "" {
		return nil, fmt.Errorf("%s header not found", HTTP_HEADER_FWD_EMAIL)
	}

	// Map REST verb from http.Request to policy actions
	requestMethod := request.Header.Get(HTTP_HEADER_ORIG_METHOD)
	if requestMethod == "" {
		return nil, fmt.Errorf("%s header not found", HTTP_HEADER_ORIG_METHOD)
	}
	var action Action
	switch requestMethod {
	case http.MethodGet:
		action = ActionGet
	case http.MethodPost:
		action = ActionCreate
	case http.MethodPut:
		action = ActionUpdate
	case http.MethodPatch:
		action = ActionUpdate
	case http.MethodDelete:
		action = ActionDelete
	}

	// Determine resource/non-resource paths from the Original URL
	// Valid resource paths are:
	// /{version}/{resource}
	// /{version}/{resource}/{resourceName|resourceID}
	//
	// Valid non-resource paths:
	// /
	// /{version}
	// /{specialPrefix} e.g /echo
	requestPath := request.Header.Get(HTTP_HEADER_ORIG_URL)
	log.Debugf("Request path: %s\n", requestPath)
	if requestPath == "" {
		return nil, fmt.Errorf("%s header not found", HTTP_HEADER_ORIG_URL)
	}
	currentParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	// Check if a nonResource path is requested
	if len(currentParts) < 2 {
		return &attributesRecord{
			userEmail:         userEmail,
			path:              requestPath,
			isResourceRequest: false,
			action:            action,
		}, nil
	}
	// Note: skipping version information in parts[0]. This can be used in the future to narrow down the request scope.
	return &attributesRecord{
		userEmail:         userEmail,
		isResourceRequest: true,
		resource:          currentParts[1],
		action:            action,
	}, nil
}
