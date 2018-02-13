///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
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

// Local returns whether running as a local test
func Local() bool {
	return isLocal
}

// EnsureLocal helps skip local tests if not local
func EnsureLocal(t *testing.T) {
	if !Local() {
		t.Skip("run with DEVLOCALTEST=1")
	}
}
