///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package dev

// NO TESTS

import (
	"os"
	"testing"
)

var isLocal = func() bool {
	v, _ := os.LookupEnv("DEVLOCALTEST")
	return v == "1"
}()

func Local() bool {
	return isLocal
}

func EnsureLocal(t *testing.T) {
	if !Local() {
		t.Skip("run with DEVLOCALTEST=1")
	}
}
