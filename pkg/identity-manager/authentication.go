///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"log"
	"net/http"
	"strings"

	gooidc "github.com/coreos/go-oidc"
	errors "github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/config"
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
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
	_, err := auth.store.Add(session)
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
