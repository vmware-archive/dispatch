///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package identitymanager

import (
	"os"
	"testing"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/config"
)

var TestConfig config.Config

func TestMain(m *testing.M) {
	TestConfig = config.LoadConfiguration("../../config.dev.json")
	os.Exit(m.Run())
}
