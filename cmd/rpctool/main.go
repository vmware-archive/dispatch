///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// +build linux,amd64

package main

import (
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/vmware/vmw-guestinfo/rpcvmx"
	"github.com/vmware/vmw-guestinfo/vmcheck"
)

var (
	set  bool
	get  bool
	fork bool
)

func init() {
	flag.BoolVar(&set, "set", false, "Sets the guestinfo.KEY with the string VALUE")
	flag.BoolVar(&get, "get", false, "Returns the config string in the guestinfo.* namespace")

	flag.Parse()
}

func main() {
	log := logrus.New().WithField("app", "rpctool")

	isVM, err := vmcheck.IsVirtualWorld()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	if !isVM {
		log.Fatalf("ERROR: not in a virtual world.")
	}

	if !set && !get && !fork {
		flag.Usage()
	}

	config := rpcvmx.NewConfig()
	if set {
		if flag.NArg() != 2 {
			log.Fatalf("ERROR: Please provide guestinfo key / value pair (eg; -set foo bar")
		}
		if err := config.SetString(flag.Arg(0), flag.Arg(1)); err != nil {
			log.Fatalf("ERROR: SetString failed with %s", err)
		}
	}

	if get {
		if flag.NArg() != 1 {
			log.Fatalf("ERROR: Please provide guestinfo key (eg; -get foo)")
		}
		if out, err := config.String(flag.Arg(0), ""); err != nil {
			log.Fatalf("ERROR: String failed with %s", err)
		} else {
			fmt.Printf("%s\n", out)
		}
	}
}
