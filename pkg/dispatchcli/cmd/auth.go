///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-openapi/runtime"
	apiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
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
	if token := viper.GetString("dispatchToken"); len(token) != 0 {
		return apiclient.BearerToken(token)
	}

	// Check if service-account & sign-key are present, gen/sign JWT token
	serviceAccount := viper.GetString("serviceAccount")
	signKeyPath := viper.GetString("jwtPrivateKey")
	if len(serviceAccount) != 0 && len(signKeyPath) != 0 {
		token, err := generateAndSignJWToken(serviceAccount, signKeyPath)
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
func generateAndSignJWToken(serviceAccount, signKeyPath string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": serviceAccount,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(jwtExpDuration).Unix(),
	})

	signBytes, err := ioutil.ReadFile(signKeyPath)
	if err != nil {
		fmt.Printf("error reading key file: %s\n", err.Error())
		return "", err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		fmt.Printf("error parsing RSA private key: %s\n", err.Error())
		return "", err
	}

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		fmt.Printf("error signing token: %s\n", err.Error())
		return "", err
	}

	return tokenString, nil
}
