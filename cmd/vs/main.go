///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package main

// NO TEST

import (
	"os"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/cmd"
)

func main() {
	cli := cmd.NewVSCLI(os.Stdin, os.Stdout, os.Stderr)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
