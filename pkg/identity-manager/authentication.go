///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"log"
	"net/http"
	"strings"
	"time"

	gooidc "github.com/coreos/go-oidc"
	errors "github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/vmware/dispatch/pkg/config"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
)

// AuthService is a service for Authentication and Security Stuff
type AuthService struct {
	Oidc  OIDC
	Csrf  *CSRF
	store entitystore.EntityStore
}

// NewAuthService is a constructor for AuthService
func NewAuthService(config config.Config, store entitystore.EntityStore) *AuthService {

	auth := new(AuthService)
	auth.Oidc = NewOIDCImpl(config)
	auth.Csrf = new(CSRF)
	auth.store = store
	return auth
}

// NewSession creates a new Session entity, based on an IDToken
func NewSession(idToken *gooidc.IDToken) *Session {

	name := strings.Replace(idToken.Subject, "@", "_", -1)
	s := Session{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           name,
		},
		IDToken: *idToken,
	}
	return &s
}

// CreateAndSaveSession creates a new session and save it into entity store
func (auth *AuthService) CreateAndSaveSession(idToken *gooidc.IDToken) (string, error) {

	session := NewSession(idToken)
	err := auth.store.Get(IdentityManagerFlags.OrgID, session.Name, session)
	if err == nil {
		// session already existes
		// return the existing one
		return session.Name, nil
	}
	_, err = auth.store.Add(session)
	if err != nil {
		return "", errors.Wrap(err, "session error")
	}
	return session.Name, nil
}

// GetSession retrieves a session from entity store
func (auth *AuthService) GetSession(id string) (*Session, error) {

	var session Session
	err := auth.store.Get(IdentityManagerFlags.OrgID, id, &session)
	if err != nil {
		return nil, errors.New("Invalid Session")
	}
	return &session, nil
}

// RemoveSession removes a session from entity store,
// Note: because there's no deletion operation at entity store
// What we actaully did is set valid flag to false.
func (auth *AuthService) RemoveSession(id string) error {

	session, err := auth.GetSession(id)
	if err != nil {
		return errors.Wrap(err, "RemoveSession Error:")
	}
	err = auth.store.Delete(IdentityManagerFlags.OrgID, session.Name, session)
	return errors.Wrap(err, "RemoveSession Error:")
}

// _DefaultCookieName is a local constant,
// used to fill the name field of a cookie
const _DefaultCookieName = "sessionId"

// NewDefaultCookie takes a string, encode it and return a http cookie
func NewDefaultCookie(id string) *http.Cookie {
	value := encodeDefaultCookie(id)
	return &http.Cookie{Name: _DefaultCookieName, Value: value, Path: "/"}
}

// ParseDefaultCookie parse and decode it, return the embeded session id
func ParseDefaultCookie(raw string) (string, error) {
	cookie, err := (&http.Request{Header: http.Header{"Cookie": {raw}}}).Cookie(_DefaultCookieName)
	if err != nil {
		return "", errors.Wrap(err, "invalid cookie")
	}
	return decodeDefaultCookie(cookie.Value), nil
}

// decodeDefaultCookie decode an encoded cookie value
func decodeDefaultCookie(value string) string {
	// TODO: add decyption algor
	return value
}

// encodeDefaultCookie encode a plaintext string
func encodeDefaultCookie(id string) string {
	// TODO: add encryption and random nouce
	return id
}

// CSRF is a service to get/verify CSRF State
type CSRF struct{}

// GetCSRFState generates a new CSRF state
func (c *CSRF) GetCSRFState() string {
	// TODO: add actual CSRF protection
	return "foobar"
}

// VerifyCSRFState verifies CSRF state
func (c *CSRF) VerifyCSRFState(state string) bool {
	// TODO: add actual CSRF protection
	return "foobar" == state
}

// Config is a interface to cover both the oauth2.Config and the mocked Config for testing
type Config interface {
	AuthCodeURL(string, ...oauth2.AuthCodeOption) string
	Exchange(context.Context, string) (*oauth2.Token, error)
}

// Verifier is a interface to cover gooidc.Verifier and mocked Verifier for testing
type Verifier interface {
	Verify(context.Context, string) (*gooidc.IDToken, error)
}

// OIDC defines functions related to OIDC
type OIDC interface {
	GetAuthEndpoint(string) string
	ExchangeIDToken(string) (*gooidc.IDToken, error)
	ExchangeIDTokenWithPassword(string, string) (*gooidc.IDToken, error)
}

// OIDCImpl is used to talk with OIDC Provider
type OIDCImpl struct {
	Config   Config
	Context  context.Context
	Verifier Verifier
}

// NewOIDCImpl creates a new OIDC instance
func NewOIDCImpl(localConfig config.Config) *OIDCImpl {

	o := new(OIDCImpl)
	o.Context = context.Background()
	provider, err := gooidc.NewProvider(o.Context, localConfig.Identity.OIDCProvider)
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &gooidc.Config{
		ClientID: localConfig.Identity.ClientID,
	}
	o.Verifier = provider.Verifier(oidcConfig)
	o.Config = &oauth2.Config{
		ClientID:     localConfig.Identity.ClientID,
		ClientSecret: localConfig.Identity.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  localConfig.Identity.RedirectURL,
		Scopes:       localConfig.Identity.Scopes,
	}
	return o
}

// GetAuthEndpoint gets auth endpoints of the OAuth2/OIDC Provider
func (o *OIDCImpl) GetAuthEndpoint(csrfState string) string {
	return o.Config.AuthCodeURL(csrfState)
}

// ExchangeIDToken uses an AuthCode to exchange for an ID_Token
func (o *OIDCImpl) ExchangeIDToken(code string) (*gooidc.IDToken, error) {

	oauth2Token, err := o.Config.Exchange(o.Context, code)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to Exchange Token")
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("No id_token field in oauth2 token")
	}
	idToken, err := o.Verifier.Verify(o.Context, rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to verify ID Token")
	}
	return idToken, nil
}

// MockedPasswordCredentialsToken is a fake function to exchange username/password to OAuth2.0 ID Token
// it should be replace in the future
func MockedPasswordCredentialsToken(c context.Context, username, password string) (*oauth2.Token, error) {

	for _, el := range config.StaticUsers.Data {
		if el.Username == username && el.Password == password {

			raw := make(map[string]interface{})
			raw["id_token"] = gooidc.IDToken{
				Issuer:   "A Fake OAuth2.0 Provider",
				Audience: []string{"A Fake OAuth2.0 Client"},
				Subject:  username,
				IssuedAt: time.Now(),
				Expiry:   time.Now().Add(time.Hour),
				Nonce:    "A Fake Nonce",
			}
			token := &oauth2.Token{
				AccessToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJkMzAxMTQxMy05NzNiLTQ1MTgtOTEwMi05MjExYTI1YWE3NWUiLCJwcm4iOiJhZG1pbkBTVkEiLCJkb21haW4iOiJMb2NhbCBVc2VycyIsInVzZXJfaWQiOiIyIiwiYXV0aF90aW1lIjoxNDM4NjQ1ODkzLCJpc3MiOiJodHRwczovL2d3LWFhLmhzLnRyY2ludC5jb20vU0FBUy9BUEkvMS4wL1JFU1QvYXV0aC90b2tlbiIsImF1ZCI6Imh0dHBzOi8vZ3ctYWEuaHMudHJjaW50LmNvbSIsImN0eCI6Ilt7XCJtdGRcIjpcInVybjpvYXNpczpuYW1lczp0YzpTQU1MOjIuMDphYzpjbGFzc2VzOlBhc3N3b3JkUHJvdGVjdGVkVHJhbnNwb3J0XCIsXCJpYXRcIjoxNDM4NjQ1ODkzLFwiaWRcIjo0fV0iLCJzY3AiOiJwcm9maWxlIGFkbWluIHVzZXIgZW1haWwiLCJpZHAiOiIwIiwiZW1sIjoiYWRtaW5Adm13YXJlLmNvbSIsImNpZCI6ImV4YW1wbGVfYnJvd3Nlcl9jbGlfY2xpZW50aWQiLCJkaWQiOiIiLCJ3aWQiOiIiLCJleHAiOjE0Mzg2Njc0OTMsImlhdCI6MTQzODY0NTg5Mywic3ViIjoiZmY5MWFiNGYtZmQ4Ny00Y2Y4LTgzZTEtOGUxMjEwOWE5Mzg4IiwicHJuX3R5cGUiOiJVU0VSIn0.YidJ6fUDxIX5uOFaGPGsmyMbg1exwwq1CrgDJ0-QoCRXZ0rbtRSFUiEjZFjQ16d5MBKUhnVeGMUYIwjC1nLDdQC_ZPVXTWg9prhcALVCSrVE52dM4OqhPdiV6aRYZjDoSEiEzAZZq1XpnPHxxze-msI6TdXCwusg35ZwTLBMcAo",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(time.Hour), // one hour
			}
			token = token.WithExtra(raw)
			return token, nil
		}
	}
	// if username == "admin" && password == "admin" {
	return nil, errors.New("Unauthorized")
}

// ExchangeIDTokenWithPassword accepts username/password, and exchange it for OIDC ID Token
func (o *OIDCImpl) ExchangeIDTokenWithPassword(username, password string) (*gooidc.IDToken, error) {

	// TODO: use the real library function instead of the faked one,
	//		 after an external vIDM instance is available
	// oauth2Token, err := o.Config.PasswordCredentialsToken(o.Context, username, password)
	oauth2Token, err := MockedPasswordCredentialsToken(o.Context, username, password)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to Exchange Password")
	}

	idToken, ok := oauth2Token.Extra("id_token").(gooidc.IDToken)
	if !ok {
		return nil, errors.New("No id_token field in oauth2 token")
	}

	// TODO: add this back when the PasswordCredentialsToken is not a faked
	// idToken, err := o.Verifier.Verify(o.Context, rawIDToken)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "Failed to verify ID Token")
	// }
	return &idToken, nil
}
