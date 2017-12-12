///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"testing"

	"github.com/vmware/dispatch/pkg/config"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/testing/store"
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
