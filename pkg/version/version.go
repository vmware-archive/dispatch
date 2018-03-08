///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package version

import (
	"fmt"
	"runtime"
)

var (
	version   string
	commit    string
	buildDate string
)

// NO TESTS

// BuildInfo describes build metadata
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

// Get returns information about the build
func Get() *BuildInfo {
	// Filled by -ldflags passed to `go build`
	return &BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
