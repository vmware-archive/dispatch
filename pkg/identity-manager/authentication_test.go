///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"testing"

	gooidc "github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/identity-manager/mocks"
	"golang.org/x/oauth2"
)

func TestSession(t *testing.T) {

	authService := GetTestAuthService(t)

	idToken1 := &gooidc.IDToken{
		Subject: "testUser1",
	}
	idToken2 := &gooidc.IDToken{
		Subject: "testUser2",
	}

	sessionID1, err := authService.CreateAndSaveSession(idToken1)
	assert.NoError(t, err, "Unexpected Error")
	sessionID2, err := authService.CreateAndSaveSession(idToken2)
	assert.NoError(t, err, "Unexpected Error")

	session1, err := authService.GetSession(sessionID1)
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, idToken1.Subject, session1.Name)
	assert.Equal(t, sessionID1, session1.Name)

	session2, err := authService.GetSession(sessionID2)
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, idToken2.Subject, session2.Name)
	assert.Equal(t, sessionID2, session2.Name)

	// remove 1
	err = authService.RemoveSession(sessionID1)
	assert.NoError(t, err, "Unexpected Error")

	session1, err = authService.GetSession(sessionID1)
	assert.Error(t, err, "Session should not exist")
	assert.Nil(t, session1)

	session2, err = authService.GetSession(sessionID2)
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, idToken2.Subject, session2.Name)
	assert.Equal(t, sessionID2, session2.Name)

	// remove 2
	err = authService.RemoveSession(sessionID2)
	assert.NoError(t, err, "Unexpected Error")

	session2, err = authService.GetSession(sessionID2)
	assert.Error(t, err, "Session should not exist")
	assert.Nil(t, session2)
}

func TestCookie(t *testing.T) {

	expectedValue1 := "testUser1Value"
	cookie1 := NewDefaultCookie(expectedValue1)

	realValue1, err := ParseDefaultCookie(cookie1.String())
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, expectedValue1, realValue1)
}

func TestCRSF(t *testing.T) {
	csrf := new(CSRF)
	assert.Equal(t, "foobar", csrf.GetCSRFState())
	assert.Equal(t, true, csrf.VerifyCSRFState("foobar"))
	assert.Equal(t, false, csrf.VerifyCSRFState("barfoo"))
}

func TestOIDC(t *testing.T) {

	o := NewOIDCImpl(*GetTestConfig(t))

	uri := o.GetAuthEndpoint("foobar")
	assert.Equal(t, "https://dev.vmwareidentity.asia/SAAS/auth/oauth2/authorize?client_id=webapp.samples.vmware.com&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Flogin%2Fvmware&response_type=code&scope=openid&state=foobar", uri)

	mockedToken := &oauth2.Token{AccessToken: "an access token"}
	raw := make(map[string]interface{})
	raw["id_token"] = "anIdToken"
	mockedToken = mockedToken.WithExtra(raw)

	mockedConfig := new(mocks.Config)
	mockedConfig.On("Exchange", o.Context, "aAuthCode").Return(mockedToken, nil)

	mockedVerifier := new(mocks.Verifier)
	mockedIDToken := &gooidc.IDToken{
		Issuer:   "FakedIssuer",
		Audience: []string{"aud1", "aud2"},
		Subject:  "testUser1"}
	mockedVerifier.On("Verify", o.Context, "anIdToken").Return(mockedIDToken, nil)

	o.Config = mockedConfig
	o.Verifier = mockedVerifier
	idToken, err := o.ExchangeIDToken("aAuthCode")
	t.Logf("error %s\n", err)
	assert.NoError(t, err, "Incorrect Auth Code: An Error Should Occur")
	assert.NotNil(t, idToken, "Id Token Should Not Be Nil")
	assert.Equal(t, mockedIDToken, idToken)
}
