///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"testing"

	gooidc "github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/mocks"
	"golang.org/x/oauth2"
)

func TestCookie(t *testing.T) {

	cs := NewCookieStore()

	expectedValue1 := "testUser1Value"
	expectedValue2 := "testUser2Value"
	cookie1 := cs.SaveCookie("testUser1", expectedValue1)
	assert.Equal(t, "username=testUser1; Path=/", cookie1.String())
	cookie2 := cs.SaveCookie("testUser2", expectedValue2)
	assert.Equal(t, "username=testUser2; Path=/", cookie2.String())

	real, err := cs.VerifyCookie("username=testUser1")
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, real, expectedValue1)

	real, err = cs.VerifyCookie("username=testUser2")
	assert.NoError(t, err, "Unexpected Error")
	assert.Equal(t, real, expectedValue2)

	cookie := cs.RemoveCookie("testUser1")
	assert.Equal(t, "", cookie.Value)
	real, err = cs.VerifyCookie("username=testUser1")
	assert.Error(t, err, "An Error Should Occur")
	assert.Nil(t, real, "It Should Be A <nil>")
}

func TestCRSF(t *testing.T) {
	csrf := new(CSRF)
	assert.Equal(t, "foobar", csrf.GetCSRFState())
	assert.Equal(t, true, csrf.VerifyCSRFState("foobar"))
	assert.Equal(t, false, csrf.VerifyCSRFState("barfoo"))
}

func TestOIDC(t *testing.T) {

	o := NewOIDCImpl(TestConfig)

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
