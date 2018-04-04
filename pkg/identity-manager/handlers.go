///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/casbin/casbin"
	jwt "github.com/dgrijalva/jwt-go"
	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	policyOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/policy"
	svcAccountOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/serviceaccount"
)

// IdentityManagerFlags are configuration flags for the identity manager
var IdentityManagerFlags = struct {
	CookieName           string `long:"cookie-name" description:"The cookie name used to identify users" default:"_oauth2_proxy"`
	SkipAuth             bool   `long:"skip-auth" description:"Skips authorization, not to be used in production env"`
	EnableBootstrapMode  bool   `long:"enable-bootstrap-mode" description:"Enabled bootstrap mode"`
	DbFile               string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend            string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser               string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword           string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase           string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	ResyncPeriod         int    `long:"resync-period" description:"The time period (in seconds) to refresh policies" default:"30"`
	OAuth2ProxyAuthURL   string `long:"oauth2-proxy-auth-url" description:"The local url for oauth2proxy service's auth endpoint'" default:"http://localhost:4180/v1/iam/oauth2/auth"`
	ServiceAccountDomain string `long:"service-account-domain" description:"The default domain name to use for service accounts" default:"svc.dispatch.local"`
	OrgID                string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
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

// HTTP constants
const (
	HTTPHeaderReqURI     = "X-Auth-Request-Redirect"
	HTTPHeaderOrigMethod = "X-Original-Method"
	HTTPHeaderFwdEmail   = "X-Auth-Request-Email"
)

// Identity manager action constants
const (
	ActionGet    Action = "get"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Action defines the type for an action
type Action string

// Identity manager resources type constants
const (
	ResourceIAM Resource = "iam"
)

// Resource defines the type for a resource
type Resource string

// Handlers defines the interface for the identity manager handlers
type Handlers struct {
	watcher  controller.Watcher
	store    entitystore.EntityStore
	enforcer *casbin.SyncedEnforcer
}

// NewHandlers create a new Policy Manager Handler
func NewHandlers(watcher controller.Watcher, store entitystore.EntityStore, enforcer *casbin.SyncedEnforcer) *Handlers {
	return &Handlers{
		watcher:  watcher,
		store:    store,
		enforcer: enforcer,
	}
}

// SetupEnforcer sets up the casbin enforcer
func SetupEnforcer(store entitystore.EntityStore) *casbin.SyncedEnforcer {
	model := casbin.NewModel(casbinPolicyModel)
	adapter := NewCasbinEntityAdapter(store)
	enforcer := casbin.NewSyncedEnforcer(model, adapter)
	return enforcer
}

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

	a.BearerAuth = func(token string) (interface{}, error) {
		// currently just return the token, auth function will take care of authentication
		return token, nil
	}

	a.RootHandler = operations.RootHandlerFunc(h.root)
	a.HomeHandler = operations.HomeHandlerFunc(h.home)
	a.AuthHandler = operations.AuthHandlerFunc(h.auth)
	a.RedirectHandler = operations.RedirectHandlerFunc(h.redirect)
	// Policy API Handlers
	a.PolicyAddPolicyHandler = policyOperations.AddPolicyHandlerFunc(h.addPolicy)
	a.PolicyGetPoliciesHandler = policyOperations.GetPoliciesHandlerFunc(h.getPolicies)
	a.PolicyGetPolicyHandler = policyOperations.GetPolicyHandlerFunc(h.getPolicy)
	a.PolicyDeletePolicyHandler = policyOperations.DeletePolicyHandlerFunc(h.deletePolicy)
	a.PolicyUpdatePolicyHandler = policyOperations.UpdatePolicyHandlerFunc(h.updatePolicy)
	// Service Account API Handlers
	a.ServiceaccountAddServiceAccountHandler = svcAccountOperations.AddServiceAccountHandlerFunc(h.addServiceAccount)
	a.ServiceaccountGetServiceAccountHandler = svcAccountOperations.GetServiceAccountHandlerFunc(h.getServiceAccount)
	a.ServiceaccountGetServiceAccountsHandler = svcAccountOperations.GetServiceAccountsHandlerFunc(h.getServiceAccounts)
	a.ServiceaccountDeleteServiceAccountHandler = svcAccountOperations.DeleteServiceAccountHandlerFunc(h.deleteServiceAccount)
	a.ServiceaccountUpdateServiceAccountHandler = svcAccountOperations.UpdateServiceAccountHandlerFunc(h.updateServiceAccount)
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

func (h *Handlers) validateAndParseToken(token string) (jwt.MapClaims, bool) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm is same as expected. This is important after the vulnerabilities with JWT using asymmetric
		// keys that don't validate the algorithm.
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// Lookup
		claims := token.Claims.(jwt.MapClaims)
		if s, ok := claims["iss"]; ok {
			unverifiedIssuer := s.(string)
			log.Debugf("Found issuer %s from unvalidated token", unverifiedIssuer)

			// Fetch service account record
			svcAccount := ServiceAccount{}
			opts := entitystore.Options{
				Filter: entitystore.FilterExists(),
			}
			if err := h.store.Get(IdentityManagerFlags.OrgID, unverifiedIssuer, opts, &svcAccount); err != nil {
				return nil, errors.Wrap(err, "store error when getting service account")
			}
			pubPEM, err := base64.StdEncoding.DecodeString(svcAccount.PublicKey)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("error when decoding public key for issuer %s", unverifiedIssuer))
			}
			block, _ := pem.Decode([]byte(pubPEM))

			if block == nil {
				return nil, errors.New("error while parsing public key: no PEM block found")
			}

			publicRSAKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, errors.Wrap(err, "Error while parsing public key")
			}
			// TODO: Validate Audience
			// TODO: Validate Token issued duration was not more than 1 hour (or min duration setting)
			return publicRSAKey, nil
		}
		// Missing issuer claim
		return nil, errors.New("missing issuer claim in unvalidated token")
	})

	if err != nil {
		log.Debugf("Error validating token: %s", err)
		return nil, false
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, true
	}
	log.Debugf("Invalid bearer token")
	return nil, false
}

func (h *Handlers) auth(params operations.AuthParams, principal interface{}) middleware.Responder {
	// For development use cases, not recommended in production env.
	if IdentityManagerFlags.SkipAuth {
		log.Warn("Skipping authorization. This is not recommended in production environments.")
		return operations.NewAuthAccepted()
	}

	// Represents a  Service Account or an User Account principle
	var subject string

	// Authenticate Account
	// Method 1: Check if bearer token exists (only service accounts are supported in this method)
	authHeader := strings.TrimSpace(params.HTTPRequest.Header.Get("Authorization"))
	if authHeader != "" {
		log.Debugf("Found Authorization header in request")
		parts := strings.Split(authHeader, " ")
		if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Debugf("Only 'Authorization: Bearer' is supported")
			return operations.NewAuthForbidden()
		}

		jwtToken := parts[1]
		if claims, ok := h.validateAndParseToken(jwtToken); ok {
			if issuer, ok := claims["iss"]; ok {
				log.Debugf("Found issuer %s from valid token", issuer)
				// Setting subject to issuer
				subject = issuer.(string)
			}
		} else {
			return operations.NewAuthForbidden()
		}
	} else {
		// Method 2: Check Cookie (only user accounts are supported in this method)
		cookie, err := params.HTTPRequest.Cookie(IdentityManagerFlags.CookieName)
		if err != nil {
			log.Debugf("Unable to find cookie in the original request: %s", err)
			return operations.NewAuthForbidden()
		}

		// Make a request to Oauth2Proxy to validate the cookie. Oauth2Proxy must be setup locally
		proxyReq, err := http.NewRequest(http.MethodGet, IdentityManagerFlags.OAuth2ProxyAuthURL, nil)
		if err != nil {
			log.Debugf("Error creating forwarding request to oauth2proxy: %s", err)
			return operations.NewAuthForbidden()
		}
		proxyReq.AddCookie(cookie)
		resp, err := http.DefaultClient.Do(proxyReq)

		if err != nil {
			log.Debugf("Error forwarding request to oauth2proxy: %s", err)
			return operations.NewAuthForbidden()
		}
		if resp.StatusCode != http.StatusAccepted {
			log.Debugf("Authentication failed with oauth2proxy: error code %v", resp.StatusCode)
			return operations.NewAuthForbidden()
		}

		// If authenticated, get subject
		log.Debugf("Received Headers from oauth2proxy %s", resp.Header)
		subject = resp.Header.Get(HTTPHeaderFwdEmail)
		if subject == "" {
			log.Debugf("Authentication Failed: Missing %s header in response from oauth2proxy", HTTPHeaderFwdEmail)
			return operations.NewAuthForbidden()
		}
	}
	// At this point, the user is authenticated, let's do a policy check.
	attrs, err := getRequestAttributes(params.HTTPRequest, subject)
	if err != nil {
		log.Debugf("Unable to parse request attributes: %s", err)
		return operations.NewAuthForbidden()
	}
	log.Debugf("Enforcing Policy: %s, %s, %s\n", attrs.subject, attrs.resource, attrs.action)

	// Skip policy check for bootstrap user.
	if IdentityManagerFlags.EnableBootstrapMode {
		log.Warn("Bootstrap mode is enabled. Please ensure it is turned off in a production environment.")
		if bootstrapUser := os.Getenv("IAM_BOOTSTRAP_USER"); bootstrapUser != "" && bootstrapUser == attrs.subject {
			// Bootstrap user can only perform on IAM resource
			if Resource(attrs.resource) != ResourceIAM {
				log.Warn("Found Bootstrap user operating on non-iam resource, auth forbidden")
				return operations.NewAuthForbidden()
			}
			log.Info("Bootstrap user auth accepted")
			return operations.NewAuthAccepted()
		}
		log.Warn("Not bootstrap user in bootstrap mode, auth forbidden")
		return operations.NewAuthForbidden()
	}

	// Note: Non-Resource requests are currently not authz enforced.
	if !attrs.isResourceRequest {
		return operations.NewAuthAccepted()
	}

	if h.enforcer.Enforce(attrs.subject, attrs.resource, string(attrs.action)) == true {
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

func getRequestAttributes(request *http.Request, subject string) (*attributesRecord, error) {
	log.Debugf("Headers: %s; Subject %s\n", request.Header, subject)

	if strings.TrimSpace(subject) == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}
	// Map REST verb from http.Request to policy actions
	requestMethod := request.Header.Get(HTTPHeaderOrigMethod)
	if requestMethod == "" {
		return nil, fmt.Errorf("%s header not found", HTTPHeaderOrigMethod)
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
	requestPath := request.Header.Get(HTTPHeaderReqURI)
	log.Debugf("Request path: %s", requestPath)
	if requestPath == "" {
		return nil, fmt.Errorf("%s header not found", HTTPHeaderReqURI)
	}
	currentParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	// Check if a nonResource path is requested
	if len(currentParts) < 2 {
		return &attributesRecord{
			subject:           subject,
			path:              requestPath,
			isResourceRequest: false,
			action:            action,
		}, nil
	}
	// Note: skipping version information in parts[0]. This can be used in the future to narrow down the request scope.
	return &attributesRecord{
		subject:           subject,
		isResourceRequest: true,
		resource:          currentParts[1],
		action:            action,
	}, nil
}
