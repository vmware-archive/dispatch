///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi/operations"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/restapi/operations/authentication"
)

// ConfigureHandlers registers the identity manager handlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, authService *AuthService) {

	a, ok := api.(*operations.IdentityManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuthAuth = func(token string) (interface{}, error) {
		return authService.CookieStore.VerifyCookie(token)
	}

	a.AuthenticationLoginHandler = authentication.LoginHandlerFunc(
		func(params authentication.LoginParams) middleware.Responder {
			// not previously logged in
			redirectURI := authService.Oidc.GetAuthEndpoint(authService.Csrf.GetCSRFState())
			return authentication.NewLoginFound().WithLocation(redirectURI)
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
			// for dev only, set cookie to the plaintext username
			principle := &models.Principle{Name: swag.String(idToken.Subject)}
			cookie := authService.CookieStore.SaveCookie(idToken.Subject, principle)
			return authentication.NewLoginVmwareFound().WithLocation("/v1/iam/home").WithSetCookie(cookie.String())
		})

	a.AuthenticationLogoutHandler = authentication.LogoutHandlerFunc(
		func(params authentication.LogoutParams, principle_ interface{}) middleware.Responder {

			principle, ok := principle_.(*models.Principle)
			if !ok {
				return authentication.NewLogoutDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{Code: http.StatusInternalServerError,
						Message: swag.String("Type Conversion Error")})
			}

			cookie := authService.CookieStore.RemoveCookie(*principle.Name)
			return authentication.NewLogoutOK().WithSetCookie(cookie.String()).WithPayload(
				&models.Message{Message: swag.String("You Have Successfully Logged Out")})
		})

	a.HomeHandler = operations.HomeHandlerFunc(
		func(params operations.HomeParams, principle_ interface{}) middleware.Responder {

			principle, ok := principle_.(*models.Principle)
			if !ok {
				return operations.NewHomeDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{Code: http.StatusInternalServerError,
						Message: swag.String("Type Conversion Error")})
			}

			// already logged in, redirect to home
			message := fmt.Sprintf("Hello %s", *principle.Name)
			return operations.NewHomeOK().WithPayload(
				&models.Message{Message: swag.String(message)})
		})

}
