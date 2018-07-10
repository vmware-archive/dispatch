///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

var (
	testOrgA = "testOrgA"
	testOrgB = "testOrgB"
)

func createTestJWT(issuer string) string {
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	if issuer != "" {
		claims["iss"] = issuer
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	pvtKeyData, _ := ioutil.ReadFile("testdata/test_key")
	pvtKey, _ := jwt.ParseRSAPrivateKeyFromPEM(pvtKeyData)
	signedToken, _ := token.SignedString(pvtKey)
	return signedToken
}

func createTestJWTHMAC(issuer string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
		"iss": issuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signedToken, _ := token.SignedString([]byte("shared_secret"))
	return signedToken
}

func addTestData(store entitystore.EntityStore) {
	// Add test policies and rules
	policy1 := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgA,
			Name:           "test-policy-1",
			Status:         entitystore.StatusREADY,
		},
		Rules: []Rule{
			{
				Subjects:  []string{"readonly-user@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"get"},
			},
			{
				Subjects:  []string{"org-admin@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			}},
	}
	policy2 := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: testOrgA,
			Name:           "test-policy-2",
			Status:         entitystore.StatusREADY,
		},
		Rules: []Rule{
			{
				Subjects:  []string{"super-admin@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			},
			{
				Subjects:  []string{"super-admin.svc.testOrgA"},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			}},
		Global: true,
	}
	store.Add(context.Background(), policy1)
	store.Add(context.Background(), policy2)
}

func setupTestAPI(t *testing.T, addTestPolicies bool) *operations.IdentityManagerAPI {
	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	o1 := &Organization{
		BaseEntity: entitystore.BaseEntity{
			Name:           testOrgA,
			OrganizationID: testOrgA,
		},
	}
	o2 := &Organization{
		BaseEntity: entitystore.BaseEntity{
			Name:           testOrgB,
			OrganizationID: testOrgB,
		},
	}
	es.Add(context.Background(), o1)
	es.Add(context.Background(), o2)
	if addTestPolicies {
		addTestData(es)
	}
	enforcer := SetupEnforcer(es)
	handlers := NewHandlers(nil, es, enforcer)
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)
	return api
}

func TestHomeHandler(t *testing.T) {
	api := setupTestAPI(t, false)

	params := operations.HomeParams{
		HTTPRequest: httptest.NewRequest("GET", "/v1/iam/home", nil),
	}
	responder := api.HomeHandler.Handle(params, nil)

	var respBody v1.Message
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)
	assert.Equal(t, "Home Page, You have already logged in", *respBody.Message)
}

func TestRootHandler(t *testing.T) {

	api := setupTestAPI(t, false)

	params := operations.RootParams{
		HTTPRequest: httptest.NewRequest("GET", "/", nil),
	}
	responder := api.RootHandler.Handle(params)

	var respBody v1.Message
	helpers.HandlerRequest(t, responder, &respBody, http.StatusOK)
	assert.Equal(t, "Default Root Page", *respBody.Message)
}

func TestAuthHandlerPolicyPass(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "readonly-user@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusAccepted)
	assert.Equal(t, testOrgA, resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthHandlerWithoutPolicyData(t *testing.T) {

	api := setupTestAPI(t, false)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "readonly-user@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusForbidden)
	assert.Equal(t, "", resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthHandlerInvalidOrgHeader(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	invalidOrg := "invalidOrg"
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &invalidOrg,
	}
	account := &authAccount{
		subject: "readonly-user@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerNonResourcePass(t *testing.T) {

	api := setupTestAPI(t, false)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/echo")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "noname@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerBootstrapPass(t *testing.T) {

	//bootstrap user can only access iam resource
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/iam/policy")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "bootstrap-user@example.com",
		kind:    subjectBootstrapUser,
	}

	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusAccepted)
	assert.Equal(t, testOrgA, resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthHandlerBootstrapInvalidOrg(t *testing.T) {

	//bootstrap user can only access iam resource
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/iam/policy")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	account := &authAccount{
		subject: "bootstrap-user@example.com",
		kind:    subjectBootstrapUser,
	}

	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusAccepted)
	assert.Equal(t, "", resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthHandlerBootstrapForbid(t *testing.T) {

	//bootstrap user can only access iam resource, will get forbidden when accessing other resources
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "bootstrap-user@example.com",
		kind:    subjectBootstrapUser,
	}

	responder := api.AuthHandler.Handle(params, account)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyFail(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "readonly-user@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyNoValidHeader(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	// Missing Req Header
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgA,
	}
	account := &authAccount{
		subject: "readonly-user@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestUserAuthOptionalOrgHeader(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	// X-Dispatch-Org header is not passed
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	account := &authAccount{
		subject:        "readonly-user@example.com",
		organizationID: testOrgA,
		kind:           subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusAccepted)
	// The subject's org is returned in the resp header
	assert.Equal(t, testOrgA, resp.Header.Get("X-Dispatch-Org"))
}

func TestUserAuthWithoutOrgHeadersMissing(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	// X-Dispatch-Org header is not passed
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// User also doesn't have OrgID after authentication
	account := &authAccount{
		subject: "readonly-user@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusForbidden)
	assert.Equal(t, "", resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthCrossOrgRequestFail(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	// X-Dispatch-Org header is not passed
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgB,
	}
	// User also doesn't have OrgID after authentication
	account := &authAccount{
		subject: "org-admin@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusForbidden)
	assert.Equal(t, "", resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthCrossOrgRequestPassUser(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	// X-Dispatch-Org header is not passed
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgB,
	}
	// User also doesn't have OrgID after authentication
	account := &authAccount{
		subject: "super-admin@example.com",
		kind:    subjectUser,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusAccepted)
	assert.Equal(t, "testOrgB", resp.Header.Get("X-Dispatch-Org"))
}

func TestAuthCrossOrgRequestSvcAccount(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	// X-Dispatch-Org header is not passed
	params := operations.AuthParams{
		HTTPRequest:  request,
		XDispatchOrg: &testOrgB,
	}
	account := &authAccount{
		subject:        "super-admin.svc.testOrgA",
		organizationID: testOrgA,
		kind:           subjectSvcAccount,
	}
	responder := api.AuthHandler.Handle(params, account)
	resp := helpers.HandlerRequestWithResponse(t, responder, nil, http.StatusAccepted)
	assert.Equal(t, "testOrgB", resp.Header.Get("X-Dispatch-Org"))
}

func TestGetRequestAttributesNoSubject(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	_, err := getRequestAttributes(request, "")
	assert.EqualError(t, err, "subject cannot be empty")
}

func TestGetRequestAttributesNoMethodHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	_, err := getRequestAttributes(request, "org-admin@example.com")
	assert.EqualError(t, err, "X-Original-Method header not found")
}

func TestGetRequestAttributesNoURLHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	_, err := getRequestAttributes(request, "org-admin@example.com")
	assert.EqualError(t, err, "X-Auth-Request-Redirect header not found")
}

func TestGetRequestAttributesValidResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	attrRecord, _ := getRequestAttributes(request, "org-admin@example.com")
	assert.Equal(t, "org-admin@example.com", attrRecord.subject)
	assert.Equal(t, ActionCreate, attrRecord.action)
	assert.Equal(t, "function", attrRecord.resource)
	assert.Equal(t, true, attrRecord.isResourceRequest)
	assert.Equal(t, "", attrRecord.path)
}

func TestGetRequestAttributesNonResourcePath(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/echo")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	attrRecord, _ := getRequestAttributes(request, "org-admin@example.com")
	assert.Equal(t, "org-admin@example.com", attrRecord.subject)
	assert.Equal(t, ActionGet, attrRecord.action)
	assert.Equal(t, "", attrRecord.resource)
	assert.Equal(t, false, attrRecord.isResourceRequest)
	assert.Equal(t, "/echo", attrRecord.path)
}

func TestGetRequestAttributesValidSubResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "v1/function/func_name/foo/bar")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	attrRecord, _ := getRequestAttributes(request, "org-admin@example.com")
	assert.Equal(t, "org-admin@example.com", attrRecord.subject)
	assert.Equal(t, ActionGet, attrRecord.action)
	assert.Equal(t, "function", attrRecord.resource)
	assert.Equal(t, true, attrRecord.isResourceRequest)
	assert.Equal(t, "", attrRecord.path)
}

func TestRedirectHandler(t *testing.T) {

	api := operations.NewIdentityManagerAPI(nil)
	handlers := &Handlers{}
	helpers.MakeAPI(t, handlers.ConfigureHandlers, api)

	request := httptest.NewRequest("GET", "/v1/iam/redirect", nil)
	testCookie := &http.Cookie{
		Name:  "_oauth2_proxy",
		Value: "testCookie",
	}
	request.AddCookie(testCookie)
	params := operations.RedirectParams{
		HTTPRequest: request,
		Redirect:    swag.String("http://redirect.com"),
	}
	responder := api.RedirectHandler.Handle(params, nil)

	w := httptest.NewRecorder()
	responder.WriteResponse(w, runtime.JSONProducer())
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)

	location, err := resp.Location()
	assert.Nil(t, err)

	expectedCookie := url.Values{
		"cookie": {testCookie.String()},
	}
	assert.Equal(t, fmt.Sprintf("http://redirect.com?%s", expectedCookie.Encode()), location.String())
}

func TestParseInvalidToken(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)
	err := h.validateJWTToken("invalid_token", nil)
	assert.EqualError(t, err, "error validating token: token contains an invalid number of segments")
}

func TestParseInvalidAlgorithm(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	token := createTestJWTHMAC("dummy_issuer")
	err := h.validateJWTToken(token, nil)
	assert.EqualError(t, err, "error validating token: unexpected signing method: HS384")
}

func TestParseMissingIssuerClaim(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	token := createTestJWT("")
	claims, err := h.getAuthAccountFromToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "missing issuer claim in unvalidated token")
}

func TestParseInvalidPublicKeyFormat(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test_svc1",
			OrganizationID: testOrgA,
		},
		PublicKey: "invalid",
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT(testOrgA + "/test_svc1")
	account, err := h.getAuthAccountFromToken(token)
	assert.Nil(t, account)
	assert.EqualError(t, err, "error while parsing public key: Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key")
}

func TestParseInvalidPublicKey(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key2.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test_svc1",
			OrganizationID: testOrgA,
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT(testOrgA + "/test_svc1")
	account, err := h.getAuthAccountFromToken(token)
	assert.Nil(t, account)
	assert.EqualError(t, err, "error validating token: crypto/rsa: verification error")
}

func TestParseNonExistingSvcAccount(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test_svc1",
			OrganizationID: testOrgA,
		},
		PublicKey: "dummy",
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT("testOrg/missing_svc_account")
	claims, err := h.getAuthAccountFromToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "store error when getting service account testOrg/missing_svc_account: error getting: no such entity")
}

func TestParseValidJWT(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test_svc1",
			OrganizationID: testOrgA,
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT(testOrgA + "/test_svc1")
	account, err := h.getAuthAccountFromToken(token)
	assert.Equal(t, "test_svc1", account.subject)
	assert.NoError(t, err)
}

func TestAuthenticateBearerInvalidHeader(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	principal, err := h.authenticateBearer("basic user:pwd")
	assert.Nil(t, principal)
	assert.EqualError(t, err, "invalid Authorization header, it must be of form 'Authorization: Bearer <token>'")
}

func TestAuthenticateBearerMissingToken(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	principal, err := h.authenticateBearer("bearer")
	assert.Nil(t, principal)
	assert.EqualError(t, err, "invalid Authorization header, it must be of form 'Authorization: Bearer <token>'")
}

func TestAuthenticateBearerPass(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test_svc1",
			OrganizationID: testOrgA,
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT(testOrgA + "/test_svc1")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Equal(t, "test_svc1", principal.(*authAccount).subject)
	assert.NoError(t, err)
}

func TestAuthenticateBearerFail(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key2.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name:           "test_svc1",
			OrganizationID: testOrgA,
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT(testOrgA + "/test_svc1")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "unable to validate bearer token: error validating token: crypto/rsa: verification error")
}

func TestBootstrapModeBearerToken(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)
	// Set bootstrap mode and public key
	h.BootstrapConfigPath = "testdata"
	token := createTestJWT("bootstrap-user@example.com")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Equal(t, "bootstrap-user@example.com", principal.(*authAccount).subject)
	assert.NoError(t, err)
	// Reset flag
	h.BootstrapConfigPath = "/bootstrap"
}

func TestBootstrapModeBearerInvalidToken(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	// Set bootstrap mode and public key
	bootstrapDir, err := ioutil.TempDir("", "test")
	defer os.RemoveAll(bootstrapDir)
	h.BootstrapConfigPath = bootstrapDir
	ioutil.WriteFile(bootstrapDir+"/bootstrap_user", []byte("test_user"), 0600)
	pubKey, _ := ioutil.ReadFile("testdata/test_key2.pub")
	ioutil.WriteFile(bootstrapDir+"/bootstrap_public_key", []byte(base64.StdEncoding.EncodeToString(pubKey)), 0600)
	token := createTestJWT("test_user")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "unable to validate bearer token: error validating token: crypto/rsa: verification error")
	// Reset flag
	h.BootstrapConfigPath = "/bootstrap"
}

func TestBootstrapModeBearerNoPubKey(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	// Set bootstrap mode and public key
	bootstrapDir, err := ioutil.TempDir("", "non_bootstrap_dir")
	defer os.RemoveAll(bootstrapDir)
	h.BootstrapConfigPath = bootstrapDir
	ioutil.WriteFile(bootstrapDir+"/bootstrap_user", []byte("test_user"), 0600)
	token := createTestJWT("test_user")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "unable to validate bearer token: missing public key in bootstrap mode")
	// Reset flag
	h.BootstrapConfigPath = "/bootstrap"
}

func TestAuthenticateCookiePass(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	testHttpserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieString := r.Header.Get("Cookie")
		assert.Equal(t, h.CookieName+"=testing", cookieString)
		w.Header().Add(HTTPHeaderEmail, "test-user1@example.com")
		w.WriteHeader(http.StatusAccepted)
	}))

	h.OAuth2ProxyAuthURL = testHttpserver.URL
	cookieString := h.CookieName + "=testing"
	principal, err := h.authenticateCookie(cookieString)
	assert.Equal(t, "test-user1@example.com", principal.(*authAccount).subject)
	assert.NoError(t, err)
}

func TestAuthenticateCookieUnauthenticated(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	testHttpserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	h.OAuth2ProxyAuthURL = testHttpserver.URL
	cookieString := h.CookieName + "=testing"
	principal, err := h.authenticateCookie(cookieString)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "authentication failed with oauth2proxy: error code 403")
}

func TestAuthenticateCookieMissingEmailHeader(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	testHttpserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	h.OAuth2ProxyAuthURL = testHttpserver.URL
	cookieString := h.CookieName + "=testing"
	principal, err := h.authenticateCookie(cookieString)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "authentication failed: missing X-Auth-Request-Email header in response from oauth2proxy")
}
