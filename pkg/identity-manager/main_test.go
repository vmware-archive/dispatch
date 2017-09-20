///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"testing"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/config"
	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/testing/store"
)

var testAuthService *AuthService
var testConfig *config.Config

func GetTestAuthService(t *testing.T) *AuthService {

	if testAuthService != nil {
		return testAuthService
	}
	testConfig := GetTestConfig(t)
	_, kv := store.MakeKVStore(t)
	testStore := entitystore.New(kv)
	return NewAuthService(*testConfig, testStore)
}

func GetTestConfig(t *testing.T) *config.Config {
	if testConfig != nil {
		return testConfig
	}
	testConfig := config.LoadConfiguration("../../config.dev.json")
	return &testConfig
}
