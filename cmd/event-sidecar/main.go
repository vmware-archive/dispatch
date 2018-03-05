///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

// NO TEST

import (
	"fmt"
	"os"

	"github.com/vmware/dispatch/pkg/event-sidecar"
)

func main() {
	cli := eventsidecar.NewCmd(os.Stdin, os.Stdout, os.Stderr)
	if err := cli.Execute(); err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
