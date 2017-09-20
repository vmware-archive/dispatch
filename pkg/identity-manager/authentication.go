///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"log"
	"net/http"
	"sync"

	gooidc "github.com/coreos/go-oidc"
	errors "github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/config"
)

// AuthService is a service for Authentication and Security Stuff
type AuthService struct {
	CookieStore *CookieStore
	Oidc        OIDC
	Csrf        *CSRF
}

// NewAuthService is a constructor for AuthService
func NewAuthService(config config.Config) *AuthService {

	auth := new(AuthService)
	auth.CookieStore = NewCookieStore()
	auth.Oidc = NewOIDCImpl(config)
	auth.Csrf = new(CSRF)
	return auth
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

// CookieStore is a storage for cookies
type CookieStore struct {
	store map[string]interface{}
	mutex sync.RWMutex
}

// NewCookieStore is a constructor for CookieStore
func NewCookieStore() *CookieStore {
	cs := new(CookieStore)
	cs.store = make(map[string]interface{})
	cs.mutex = sync.RWMutex{}
	return cs
}

// SaveCookie saves a cookie into the cookie store
func (cs *CookieStore) SaveCookie(key string, value interface{}) *http.Cookie {
	cookie := &http.Cookie{Name: "username", Value: key, Path: "/"}
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.store[key] = value
	return cookie
}

// VerifyCookie verifies a cookie from the client
func (cs *CookieStore) VerifyCookie(token string) (interface{}, error) {

	cookie, err := (&http.Request{Header: http.Header{"Cookie": {token}}}).Cookie("username")
	if err != nil {
		return nil, errors.Wrap(err, "No Cookie Found")
	}
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	value, ok := cs.store[cookie.Value]
	if ok == false {
		return nil, errors.New("Invalid Cookie")
	}
	return value, nil
}

// RemoveCookie removes a cookie from cookie store
func (cs *CookieStore) RemoveCookie(key string) *http.Cookie {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	delete(cs.store, key)
	cookie := &http.Cookie{Name: "username", Value: "", Path: "/"}
	return cookie
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
