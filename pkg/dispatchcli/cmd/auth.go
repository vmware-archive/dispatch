///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io/ioutil"
	"time"

	"crypto/rsa"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-openapi/runtime"
	apiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
)

const (
	jwtExpDuration = time.Duration(1) * time.Hour // expiration duration for JWT token
)

func multiAuth(writers ...runtime.ClientAuthInfoWriter) runtime.ClientAuthInfoWriter {
	return runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, registry strfmt.Registry) error {
		for _, w := range writers {
			err := w.AuthenticateRequest(r, registry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAuthInfoWriter constructor an ClientAuthInfoWriter
func GetAuthInfoWriter() runtime.ClientAuthInfoWriter {
	// Check if a jwt token was present, if true, return a Header with Bearer token
	if token := viperCtx.GetString("dispatchToken"); len(token) != 0 {
		return apiclient.BearerToken(token)
	}
	if dispatchConfig.Token != "" {
		return apiclient.BearerToken(dispatchConfig.Token)
	}

	// Check if service-account & sign-key are present, gen/sign JWT token
	serviceAccount := viperCtx.GetString("serviceAccount")
	signKeyPath := viperCtx.GetString("jwtPrivateKey")
	if len(serviceAccount) != 0 && len(signKeyPath) != 0 {
		token, err := generateAndSignJWToken(serviceAccount, nil, &signKeyPath)
		if err != nil {
			fmt.Printf("error generating JWT: %s\n", err.Error())
		}
		return apiclient.BearerToken(token)
	}
	if dispatchConfig.ServiceAccount != "" && dispatchConfig.JWTPrivateKey != "" {
		fmt.Printf("Generating JWT with %s\n", dispatchConfig.ServiceAccount)
		token, err := generateAndSignJWToken(dispatchConfig.ServiceAccount, nil, &dispatchConfig.JWTPrivateKey)
		if err != nil {
			fmt.Printf("error generating JWT: %s\n", err.Error())
		}
		return apiclient.BearerToken(token)
	}

	// Oauth2Proxy always expects a cookie header even if the server is setup with SkipAuth. Hence, set a dummy default.
	cookie := "unset"
	if dispatchConfig.Cookie != "" {
		cookie = dispatchConfig.Cookie
	}
	return apiclient.APIKeyAuth("cookie", "header", cookie)
}

// Generate and sign JWT,
func generateAndSignJWToken(serviceAccount string, rsaPvtKey *rsa.PrivateKey, pemKeyPath *string) (string, error) {

	if pemKeyPath != nil {
		signBytes, err := ioutil.ReadFile(*pemKeyPath)
		if err != nil {
			fmt.Printf("error reading key file: %s\n", err.Error())
			return "", err
		}
		rsaPvtKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
		if err != nil {
			fmt.Printf("error parsing RSA private key from pem: %s\n", err.Error())
			return "", err
		}
	}

	if rsaPvtKey == nil {
		return "", errors.New("either rsa pvt key or path to pem encoded file should be provided")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": serviceAccount,
		// Handle clock skew on the server side
		"iat": time.Now().Add(-time.Minute).Unix(),
		"exp": time.Now().Add(jwtExpDuration).Unix(),
	})

	tokenString, err := token.SignedString(rsaPvtKey)
	if err != nil {
		fmt.Printf("error signing token: %s\n", err.Error())
		return "", err
	}

	return tokenString, nil
}
