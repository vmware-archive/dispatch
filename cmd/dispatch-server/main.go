///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

// NO TEST

import (
	"os"

	// The following blank import is to load GKE auth plugin required when authenticating against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// The following blank import is to load OIDC auth plugin required when authenticating against OIDC-enabled clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/vmware/dispatch/pkg/dispatchserver"
)

func main() {
	cli := dispatchserver.NewCLI(os.Stdout)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
