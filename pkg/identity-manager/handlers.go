///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi/operations"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi/operations/authentication"
)

// IdentityManagerFlags are configuration flags for the identity manager
var IdentityManagerFlags = struct {
	StaticUsers string `long:"static-users" description:"Path to static user json file" default:"./user.dev.json"`
	Config      string `long:"config" description:"Path to configuration json file" default:"./config.dev.json"`
	DbFile      string `long:"db-file" description:"Path to BoltDB file" default:"./db.bolt"`
	OrgID       string `long:"organization" description:"(temporary) Static organization id" default:"serverless"`
}{}

func sessionEntityToModel(s *Session) *models.Session {
	m := models.Session{
		ID:   swag.String(s.ID),
		Name: swag.String(s.Name),
	}
	return &m
}

type Handlers struct {
	AuthService *AuthService
}

// ConfigureHandlers registers the identity manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {

	a, ok := api.(*operations.IdentityManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuth = func(token string) (interface{}, error) {

		sessionID, err := ParseDefaultCookie(token)
		if err != nil {
			return nil, err
		}
		session, err := h.AuthService.GetSession(sessionID)
		if err != nil {
			return nil, err
		}
		sessionModel := sessionEntityToModel(session)
		return sessionModel, nil
	}

	a.AuthenticationLoginHandler = authentication.LoginHandlerFunc(h.login)
	a.AuthenticationLoginPasswordHandler = authentication.LoginPasswordHandlerFunc(h.loginPassword)
	a.AuthenticationLoginVmwareHandler = authentication.LoginVmwareHandlerFunc(h.loginVmware)
	a.AuthenticationLogoutHandler = authentication.LogoutHandlerFunc(h.logout)
	a.HomeHandler = operations.HomeHandlerFunc(h.home)
	a.RootHandler = operations.RootHandlerFunc(h.root)
	a.RedirectHandler = operations.RedirectHandlerFunc(h.redirect)
}

func (h *Handlers) login(params authentication.LoginParams) middleware.Responder {
	// not previously logged in
	redirectURI := h.AuthService.Oidc.GetAuthEndpoint(h.AuthService.Csrf.GetCSRFState())
	return authentication.NewLoginFound().WithLocation(redirectURI)
}

func (h *Handlers) loginPassword(params authentication.LoginPasswordParams) middleware.Responder {

	username := *params.Username
	password := *params.Password
	idToken, err := h.AuthService.Oidc.ExchangeIDTokenWithPassword(username, password)
	if err != nil {
		return authentication.NewLoginPasswordDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{Code: http.StatusInternalServerError, Message: swag.String(fmt.Sprintln(err))})
	}

	sessionID, err := h.AuthService.CreateAndSaveSession(idToken)
	if err != nil {
		return authentication.NewLoginPasswordDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(fmt.Sprintln(err)),
			})
	}
	rawCookie := NewDefaultCookie(sessionID).String()
	log.Printf("login request: logged in, cookie: %s\n", rawCookie)
	return authentication.NewLoginPasswordOK().WithSetCookie(rawCookie).WithPayload(
		&models.Auth{
			Cookie: rawCookie,
		})
}

func (h *Handlers) loginVmware(params authentication.LoginVmwareParams) middleware.Responder {

	if !h.AuthService.Csrf.VerifyCSRFState(*params.State) {
		return authentication.NewLoginVmwareDefault(http.StatusBadRequest)
	}

	idToken, err := h.AuthService.Oidc.ExchangeIDToken(*params.Code)
	if err != nil {
		return authentication.NewLoginVmwareDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{Code: http.StatusInternalServerError, Message: swag.String(fmt.Sprintln(err))})
	}

	sessionID, err := h.AuthService.CreateAndSaveSession(idToken)
	if err != nil {
		return authentication.NewLoginVmwareDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(fmt.Sprintln(err)),
			})
	}
	rawCookie := NewDefaultCookie(sessionID).String()
	return authentication.NewLoginVmwareFound().WithLocation("/v1/iam/home").WithSetCookie(rawCookie)
}

func (h *Handlers) logout(params authentication.LogoutParams, sessionInterface interface{}) middleware.Responder {

	session, ok := sessionInterface.(*models.Session)
	if !ok {
		return authentication.NewLogoutDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("Type Conversion Error"),
			})
	}

	err := h.AuthService.RemoveSession(*session.Name)
	if err != nil {
		return authentication.NewLogoutDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String(err.Error()),
			})
	}
	return authentication.NewLogoutOK().WithSetCookie(NewDefaultCookie("").String()).WithPayload(
		&models.Message{Message: swag.String("You Have Successfully Logged Out")})
}

func (h *Handlers) home(params operations.HomeParams) middleware.Responder {

	message := "Home Page, You have already logged in"
	return operations.NewHomeOK().WithPayload(
		&models.Message{Message: swag.String(message)})
}

func (h *Handlers) root(params operations.RootParams) middleware.Responder {
	message := "Default Root Page, no authentication required"
	return operations.NewRootOK().WithPayload(
		&models.Message{Message: swag.String(message)})
}

func (h *Handlers) redirect(params operations.RedirectParams) middleware.Responder {

	redirect := *params.Redirect
	cookie, err := params.HTTPRequest.Cookie("_oauth2_proxy")
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
