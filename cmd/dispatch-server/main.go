///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

// NO TEST

import (
	"os"

	"github.com/vmware/dispatch/pkg/dispatchserver"
)

func main() {
	cli := dispatchserver.NewCLI(os.Stdout)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
