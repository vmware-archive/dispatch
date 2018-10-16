///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// +build linux,amd64

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vmware/vmw-guestinfo/rpcvmx"
	"github.com/vmware/vmw-guestinfo/vmcheck"

	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/xml"
)

var (
	set string
	key string
)

func init() {
	flag.StringVar(&set, "set", "", "Set value for OVF property")
	flag.StringVar(&key, "key", "", "Work on single OVF property")

	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

func run() error {
	// Check if we're running inside a VM
	isVM, err := vmcheck.IsVirtualWorld()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	}

	// No point in running if we're not inside a VM
	if !isVM {
		fmt.Println("not living in a virtual world... :(")
		return err
	}

	config := rpcvmx.NewConfig()
	var e ovf.Env

	if err := fetchovfEnv(config, &e); err != nil {
		return err
	}

	// If set and key are populated, let's set the key to the value passed
	if set != "" && key != "" {

		var props []ovf.EnvProperty

		for _, p := range e.Property.Properties {
			if p.Key == key {
				props = append(props, ovf.EnvProperty{
					Key:   p.Key,
					Value: set,
				})
			} else {
				props = append(props, ovf.EnvProperty{
					Key:   p.Key,
					Value: p.Value,
				})
			}
		}

		env := ovf.Env{
			EsxID: e.EsxID,
			Platform: &ovf.PlatformSection{
				Kind:    e.Platform.Kind,
				Version: e.Platform.Version,
				Vendor:  e.Platform.Vendor,
				Locale:  e.Platform.Locale,
			},
			Property: &ovf.PropertySection{
				Properties: props,
			},
		}
		// Send updated ovfEnv through the rpcvmx channel
		if err := config.SetString("guestinfo.ovfEnv", env.MarshalManual()); err != nil {
			return err
		}
		// Refresh ovfEnv
		if err := fetchovfEnv(config, &e); err != nil {
			return err
		}

	}

	// LET'S HAVE A MAP! SO YOU CAN DO LOOKUPS!
	m := make(map[string]string)
	for _, v := range e.Property.Properties {
		m[v.Key] = v.Value
	}

	// If a key is all we want...
	if key != "" {
		fmt.Println(m[key])
		return nil
	}

	// Let's print the whole property list by default
	for k, v := range m {
		fmt.Printf("[%s]=%s\n", k, v)
	}

	return nil
}

func fetchovfEnv(config *rpcvmx.Config, e *ovf.Env) error {
	ovfEnv, err := config.String("guestinfo.ovfEnv", "")
	if err != nil {
		return fmt.Errorf("impossible to fetch ovf environment, exiting")
	}

	if err = xml.Unmarshal([]byte(ovfEnv), &e); err != nil {
		return fmt.Errorf("error: %s", err.Error())
	}

	return nil
}
