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
	e := &Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
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
				Subjects:  []string{"super-admin@example.com"},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			}},
	}
	store.Add(context.Background(), e)
}

func setupTestAPI(t *testing.T, addTestPolicies bool) *operations.IdentityManagerAPI {
	api := operations.NewIdentityManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
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
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, "readonly-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerWithoutPolicyData(t *testing.T) {

	api := setupTestAPI(t, false)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, "readonly-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerNonResourcePass(t *testing.T) {

	api := setupTestAPI(t, false)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/echo")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, "noname@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
}

func TestAuthHandlerBootstrapPass(t *testing.T) {

	//bootstrap user can only access iam resource
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/iam/policy")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode and user
	IdentityManagerFlags.BootstrapConfigPath = "testdata"

	responder := api.AuthHandler.Handle(params, "bootstrap-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusAccepted)
	// Reset flag
	IdentityManagerFlags.BootstrapConfigPath = "/bootstrap"
}

func TestAuthHandlerBootstrapForbid(t *testing.T) {

	//bootstrap user can only access iam resource, will get forbidden when accessing other resources
	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode and user
	IdentityManagerFlags.BootstrapConfigPath = "testdata"
	responder := api.AuthHandler.Handle(params, "bootstrap-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
	// Reset flag
	IdentityManagerFlags.BootstrapConfigPath = "/bootstrap"
}

func TestAuthHandlerNonBootstrapUserInBootstrapMode(t *testing.T) {

	//non-bootstrap user in bootstrap mode cannot access any resources
	api := setupTestAPI(t, true)

	// try access iam resources
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/iam/policy")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	// Set bootstrap mode and user
	IdentityManagerFlags.BootstrapConfigPath = "testdata"
	responder := api.AuthHandler.Handle(params, "non-bootstrap-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)

	// try access non-iam resources
	request.Header.Set(HTTPHeaderReqURI, "v1/image")
	responder = api.AuthHandler.Handle(params, "non-bootstrap-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
	// Reset flag
	IdentityManagerFlags.BootstrapConfigPath = "/bootstrap"
}

func TestAuthHandlerPolicyFail(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, "readonly-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
}

func TestAuthHandlerPolicyNoValidHeader(t *testing.T) {

	api := setupTestAPI(t, true)
	request := httptest.NewRequest("GET", "/auth", nil)
	// Missing Req Header
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	params := operations.AuthParams{
		HTTPRequest: request,
	}
	responder := api.AuthHandler.Handle(params, "readonly-user@example.com")
	helpers.HandlerRequest(t, responder, nil, http.StatusForbidden)
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
	_, err := getRequestAttributes(request, "super-admin@example.com")
	assert.EqualError(t, err, "X-Original-Method header not found")
}

func TestGetRequestAttributesNoURLHeader(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	_, err := getRequestAttributes(request, "super-admin@example.com")
	assert.EqualError(t, err, "X-Auth-Request-Redirect header not found")
}

func TestGetRequestAttributesValidResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/v1/function")
	request.Header.Add(HTTPHeaderOrigMethod, "POST")
	attrRecord, _ := getRequestAttributes(request, "super-admin@example.com")
	assert.Equal(t, "super-admin@example.com", attrRecord.subject)
	assert.Equal(t, ActionCreate, attrRecord.action)
	assert.Equal(t, "function", attrRecord.resource)
	assert.Equal(t, true, attrRecord.isResourceRequest)
	assert.Equal(t, "", attrRecord.path)
}

func TestGetRequestAttributesNonResourcePath(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "/echo")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	attrRecord, _ := getRequestAttributes(request, "super-admin@example.com")
	assert.Equal(t, "super-admin@example.com", attrRecord.subject)
	assert.Equal(t, ActionGet, attrRecord.action)
	assert.Equal(t, "", attrRecord.resource)
	assert.Equal(t, false, attrRecord.isResourceRequest)
	assert.Equal(t, "/echo", attrRecord.path)
}

func TestGetRequestAttributesValidSubResource(t *testing.T) {

	request := httptest.NewRequest("GET", "/auth", nil)
	request.Header.Add(HTTPHeaderReqURI, "v1/function/func_name/foo/bar")
	request.Header.Add(HTTPHeaderOrigMethod, "GET")
	attrRecord, _ := getRequestAttributes(request, "super-admin@example.com")
	assert.Equal(t, "super-admin@example.com", attrRecord.subject)
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
	claims, err := h.parseAndValidateToken("invalid_token")
	assert.Nil(t, claims)
	assert.EqualError(t, err, "error validating token: token contains an invalid number of segments")
}

func TestParseInvalidAlgorithm(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	token := createTestJWTHMAC("dummy_issuer")
	claims, err := h.parseAndValidateToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "error validating token: unexpected signing method: HS384")
}

func TestParseMissingIssuerClaim(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	token := createTestJWT("")
	claims, err := h.parseAndValidateToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "error validating token: missing issuer claim in unvalidated token")
}

func TestParseInvalidPublicKeyFormat(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name: "test_svc1",
		},
		PublicKey: "invalid",
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT("test_svc1")
	claims, err := h.parseAndValidateToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "error validating token: error while parsing public key: Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key")
}

func TestParseInvalidPublicKey(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key2.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name: "test_svc1",
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT("test_svc1")
	claims, err := h.parseAndValidateToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "error validating token: crypto/rsa: verification error")
}

func TestParseNonExistingSvcAccount(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	token := createTestJWT("missing_svc_account")
	claims, err := h.parseAndValidateToken(token)
	assert.Nil(t, claims)
	assert.EqualError(t, err, "error validating token: store error when getting service account missing_svc_account: error getting: no such entity")
}

func TestParseValidJWT(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name: "test_svc1",
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT("test_svc1")
	claims, err := h.parseAndValidateToken(token)
	assert.Equal(t, "test_svc1", claims["iss"])
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
			Name: "test_svc1",
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT("test_svc1")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Equal(t, "test_svc1", principal)
	assert.NoError(t, err)
}

func TestAuthenticateBearerFail(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	pubKey, _ := ioutil.ReadFile("testdata/test_key2.pub")
	svcAccount := &ServiceAccount{
		BaseEntity: entitystore.BaseEntity{
			Name: "test_svc1",
		},
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
	}
	es.Add(context.Background(), svcAccount)

	token := createTestJWT("test_svc1")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "unable to validate bearer token: error validating token: crypto/rsa: verification error")
}

func TestBootstrapModeBearerToken(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)
	// Set bootstrap mode and public key
	IdentityManagerFlags.BootstrapConfigPath = "testdata"
	token := createTestJWT("bootstrap-user@example.com")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Equal(t, "bootstrap-user@example.com", principal)
	assert.NoError(t, err)
	// Reset flag
	IdentityManagerFlags.BootstrapConfigPath = "/bootstrap"
}

func TestBootstrapModeBearerInvalidToken(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	// Set bootstrap mode and public key
	bootstrapDir, err := ioutil.TempDir("", "test")
	defer os.RemoveAll(bootstrapDir)
	IdentityManagerFlags.BootstrapConfigPath = bootstrapDir
	ioutil.WriteFile(bootstrapDir+"/bootstrap_user", []byte("test_user"), 0600)
	pubKey, _ := ioutil.ReadFile("testdata/test_key2.pub")
	ioutil.WriteFile(bootstrapDir+"/bootstrap_public_key", []byte(base64.StdEncoding.EncodeToString(pubKey)), 0600)
	token := createTestJWT("bootstrap-user@example.com")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "unable to validate bearer token: error validating token: crypto/rsa: verification error")
	// Reset flag
	IdentityManagerFlags.BootstrapConfigPath = "/bootstrap"
}

func TestBootstrapModeBearerNoPubKey(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	// Set bootstrap mode and public key
	bootstrapDir, err := ioutil.TempDir("", "non_bootstrap_dir")
	defer os.RemoveAll(bootstrapDir)
	IdentityManagerFlags.BootstrapConfigPath = bootstrapDir
	ioutil.WriteFile(bootstrapDir+"/bootstrap_user", []byte("test_user"), 0600)
	token := createTestJWT("bootstrap-user@example.com")
	principal, err := h.authenticateBearer("bearer " + token)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "unable to validate bearer token: error validating token: missing public key in bootstrap mode")
	// Reset flag
	IdentityManagerFlags.BootstrapConfigPath = "/bootstrap"
}

func TestAuthenticateCookiePass(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	testHttpserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieString := r.Header.Get("Cookie")
		assert.Equal(t, IdentityManagerFlags.CookieName+"=testing", cookieString)
		w.Header().Add(HTTPHeaderEmail, "test-user1@example.com")
		w.WriteHeader(http.StatusAccepted)
	}))

	IdentityManagerFlags.OAuth2ProxyAuthURL = testHttpserver.URL
	cookieString := IdentityManagerFlags.CookieName + "=testing"
	principal, err := h.authenticateCookie(cookieString)
	assert.Equal(t, "test-user1@example.com", principal)
	assert.NoError(t, err)
}

func TestAuthenticateCookieUnauthenticated(t *testing.T) {
	es := helpers.MakeEntityStore(t)
	enforcer := SetupEnforcer(es)
	h := NewHandlers(nil, es, enforcer)

	testHttpserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	IdentityManagerFlags.OAuth2ProxyAuthURL = testHttpserver.URL
	cookieString := IdentityManagerFlags.CookieName + "=testing"
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

	IdentityManagerFlags.OAuth2ProxyAuthURL = testHttpserver.URL
	cookieString := IdentityManagerFlags.CookieName + "=testing"
	principal, err := h.authenticateCookie(cookieString)
	assert.Nil(t, principal)
	assert.EqualError(t, err, "authentication failed: missing X-Auth-Request-Email header in response from oauth2proxy")
}
