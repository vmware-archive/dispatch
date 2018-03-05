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
	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"os"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	policyOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/policy"
	"github.com/vmware/dispatch/pkg/trace"
)

// IdentityManagerFlags are configuration flags for the identity manager
var IdentityManagerFlags = struct {
	CookieName          string `long:"cookie-name" description:"The cookie name used to identify users" default:"_oauth2_proxy"`
	SkipAuth            bool   `long:"skip-auth" description:"Skips authorization, not to be used in production env"`
	EnableBootstrapMode bool   `long:"enable-bootstrap-mode" description:"Enabled bootstrap mode"`
	DbFile              string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend           string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser              string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword          string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase          string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	ResyncPeriod        int    `long:"resync-period" description:"The time period (in seconds) to refresh policies" default:"30"`
	OrgID               string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
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
	HTTPHeaderFwdEmail   = "X-Forwarded-Email"
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

	a.RootHandler = operations.RootHandlerFunc(h.root)
	a.HomeHandler = operations.HomeHandlerFunc(h.home)
	a.AuthHandler = operations.AuthHandlerFunc(h.auth)
	a.RedirectHandler = operations.RedirectHandlerFunc(h.redirect)
	// Policy API Handlers
	a.PolicyAddPolicyHandler = policyOperations.AddPolicyHandlerFunc(h.addPolicy)
	a.PolicyGetPoliciesHandler = policyOperations.GetPoliciesHandlerFunc(h.getPolicies)
	a.PolicyDeletePolicyHandler = policyOperations.DeletePolicyHandlerFunc(h.deletePolicy)
}

func policyModelToEntity(m *models.Policy) *Policy {
	defer trace.Tracef("name '%s'", *m.Name)()

	e := Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           *m.Name,
		},
	}
	for _, r := range m.Rules {
		rule := Rule{
			Subjects:  r.Subjects,
			Resources: r.Resources,
			Actions:   r.Actions,
		}
		e.Rules = append(e.Rules, rule)
	}
	return &e
}

func policyEntityToModel(e *Policy) *models.Policy {
	defer trace.Tracef("name '%s'", e.Name)()
	m := models.Policy{
		ID:           strfmt.UUID(e.ID),
		Name:         swag.String(e.Name),
		Status:       models.Status(e.Status),
		CreatedTime:  e.CreatedTime.Unix(),
		ModifiedTime: e.ModifiedTime.Unix(),
	}
	for _, r := range e.Rules {
		rule := models.Rule{
			Subjects:  r.Subjects,
			Resources: r.Resources,
			Actions:   r.Actions,
		}
		m.Rules = append(m.Rules, &rule)
	}
	return &m
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
	attrs, err := getRequestAttributes(params.HTTPRequest)
	if err != nil {
		log.Debugf("Unable to parse request attributes: %s", err)
		return operations.NewAuthForbidden()
	}
	log.Debugf("Enforcing Policy: %s, %s, %s\n", attrs.userEmail, attrs.resource, attrs.action)

	// Skip policy check for bootstrap user.
	if IdentityManagerFlags.EnableBootstrapMode {
		log.Warn("Bootstrap mode is enabled. Please ensure it is turned off in a production environment.")
		if bootstrapUser := os.Getenv("IAM_BOOTSTRAP_USER"); bootstrapUser != "" && bootstrapUser == attrs.userEmail {
			log.Warn("Found Bootstrap user, skipping policy check")
			return operations.NewAuthAccepted()
		}
	}

	// Note: Non-Resource requests are currently not authz enforced.
	if !attrs.isResourceRequest {
		return operations.NewAuthAccepted()
	}

	if h.enforcer.Enforce(attrs.userEmail, attrs.resource, string(attrs.action)) == true {
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

func (h *Handlers) getPolicies(params policyOperations.GetPoliciesParams, principal interface{}) middleware.Responder {

	defer trace.Trace("")()
	var policies []*Policy

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err := h.store.List(IdentityManagerFlags.OrgID, opts, &policies)
	if err != nil {
		log.Errorf("store error when listing policies: %+v", err)
		return policyOperations.NewGetPoliciesInternalServerError().WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting policies"),
			})
	}
	var policyModels []*models.Policy
	for _, policy := range policies {
		policyModels = append(policyModels, policyEntityToModel(policy))
	}
	return policyOperations.NewGetPoliciesOK().WithPayload(policyModels)
}

func (h *Handlers) addPolicy(params policyOperations.AddPolicyParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	policyRequest := params.Body
	e := policyModelToEntity(policyRequest)
	for _, rule := range e.Rules {
		// Do some basic validation although this must be handled at the goswagger server.
		if rule.Subjects == nil || rule.Actions == nil || rule.Resources == nil {
			return policyOperations.NewAddPolicyBadRequest().WithPayload(
				&models.Error{
					Code:    http.StatusBadRequest,
					Message: swag.String("invalid rule definition, missing required fields"),
				})
		}
	}

	e.Status = entitystore.StatusCREATING

	if _, err := h.store.Add(e); err != nil {
		log.Errorf("store error when adding a new policy %s: %+v", e.Name, err)
		return policyOperations.NewAddPolicyInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing new policy"),
		})
	}

	h.watcher.OnAction(e)

	return policyOperations.NewAddPolicyCreated().WithPayload(policyEntityToModel(e))
}

func (h *Handlers) deletePolicy(params policyOperations.DeletePolicyParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.PolicyName)()
	name := params.PolicyName

	var e Policy
	if err := h.store.Get(IdentityManagerFlags.OrgID, name, entitystore.Options{}, &e); err != nil {
		log.Errorf("store error when getting policy: %+v", err)
		return policyOperations.NewDeletePolicyNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("policy not found"),
			})
	}

	if e.Status == entitystore.StatusDELETING {
		log.Warnf("Attempting to delete policy  %s which already is in DELETING state: %+v", e.Name)
		return policyOperations.NewDeletePolicyBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Unable to delete policy %s: policy is already being deleted", e.Name)),
		})
	}

	e.Status = entitystore.StatusDELETING
	if _, err := h.store.Update(e.Revision, &e); err != nil {
		log.Errorf("store error when deleting a policy %s: %+v", e.Name, err)
		return policyOperations.NewDeletePolicyInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting a policy"),
		})
	}

	h.watcher.OnAction(&e)

	return policyOperations.NewDeletePolicyOK().WithPayload(policyEntityToModel(&e))
}

func getRequestAttributes(request *http.Request) (*attributesRecord, error) {
	log.Debugf("Headers: %s\n", request.Header)

	// Get User Info
	userEmail := request.Header.Get(HTTPHeaderFwdEmail)
	if userEmail == "" {
		return nil, fmt.Errorf("%s header not found", HTTPHeaderFwdEmail)
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
