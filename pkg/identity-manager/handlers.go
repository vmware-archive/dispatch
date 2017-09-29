///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"
	"log"
	"net/http"

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

// ConfigureHandlers registers the identity manager handlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, authService *AuthService) {

	a, ok := api.(*operations.IdentityManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuthAuth = func(token string) (interface{}, error) {

		sessionID, err := ParseDefaultCookie(token)
		if err != nil {
			return nil, err
		}
		session, err := authService.GetSession(sessionID)
		if err != nil {
			return nil, err
		}
		sessionModel := sessionEntityToModel(session)
		return sessionModel, nil
	}

	a.AuthenticationLoginHandler = authentication.LoginHandlerFunc(
		func(params authentication.LoginParams) middleware.Responder {
			// not previously logged in
			redirectURI := authService.Oidc.GetAuthEndpoint(authService.Csrf.GetCSRFState())
			return authentication.NewLoginFound().WithLocation(redirectURI)
		})

	a.AuthenticationLoginPasswordHandler = authentication.LoginPasswordHandlerFunc(
		func(params authentication.LoginPasswordParams) middleware.Responder {

			log.Printf("login request: %s\n", params.HTTPRequest.URL)
			username := *params.Username
			password := *params.Password
			log.Printf("login request: username=%s, password=%s\n", username, password)
			idToken, err := authService.Oidc.ExchangeIDTokenWithPassword(username, password)
			if err != nil {
				return authentication.NewLoginPasswordDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{Code: http.StatusInternalServerError, Message: swag.String(fmt.Sprintln(err))})
			}

			sessionID, err := authService.CreateAndSaveSession(idToken)
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
		})

	a.AuthenticationLoginVmwareHandler = authentication.LoginVmwareHandlerFunc(
		func(params authentication.LoginVmwareParams) middleware.Responder {

			if !authService.Csrf.VerifyCSRFState(*params.State) {
				return authentication.NewLoginVmwareDefault(http.StatusBadRequest)
			}

			idToken, err := authService.Oidc.ExchangeIDToken(*params.Code)
			if err != nil {
				return authentication.NewLoginVmwareDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{Code: http.StatusInternalServerError, Message: swag.String(fmt.Sprintln(err))})
			}

			sessionID, err := authService.CreateAndSaveSession(idToken)
			if err != nil {
				return authentication.NewLoginVmwareDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{
						Code:    http.StatusInternalServerError,
						Message: swag.String(fmt.Sprintln(err)),
					})
			}
			rawCookie := NewDefaultCookie(sessionID).String()
			return authentication.NewLoginVmwareFound().WithLocation("/v1/iam/home").WithSetCookie(rawCookie)
		})

	a.AuthenticationLogoutHandler = authentication.LogoutHandlerFunc(
		func(params authentication.LogoutParams, sessionInterface interface{}) middleware.Responder {

			session, ok := sessionInterface.(*models.Session)
			if !ok {
				return authentication.NewLogoutDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{
						Code:    http.StatusInternalServerError,
						Message: swag.String("Type Conversion Error"),
					})
			}

			err := authService.RemoveSession(*session.Name)
			if err != nil {
				return authentication.NewLogoutDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{
						Code:    http.StatusInternalServerError,
						Message: swag.String(err.Error()),
					})
			}
			return authentication.NewLogoutOK().WithSetCookie(NewDefaultCookie("").String()).WithPayload(
				&models.Message{Message: swag.String("You Have Successfully Logged Out")})
		})

	a.HomeHandler = operations.HomeHandlerFunc(
		func(params operations.HomeParams, session_ interface{}) middleware.Responder {

			session, ok := session_.(*models.Session)
			if !ok {
				return operations.NewHomeDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{Code: http.StatusInternalServerError,
						Message: swag.String("Type Conversion Error")})
			}

			// already logged in, redirect to home
			message := fmt.Sprintf("Hello %s", *session.Name)
			return operations.NewHomeOK().WithPayload(
				&models.Message{Message: swag.String(message)})
		})

}
