///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package version

import (
	"fmt"
	"runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// Filled by -ldflags passed to `go build`
var (
	version   string
	commit    string
	buildDate string
)

// NO TESTS

// Get returns information about the version/build
func Get() *v1.Version {
	return &v1.Version{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
