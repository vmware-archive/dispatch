///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vmware/dispatch/pkg/event-driver/drivers/vcenter"
)

// NO TESTS

// NewCmdVCenter creates a command object for vCenter driver.
func NewCmdVCenter(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcenter",
		Short: "Run vCenter driver",
		RunE:  vCenterDriverCmd(out, errOut),
	}
	cmd.Flags().String("vcenterurl", "https://vcenter.corp.local:443", "URL to vCenter instance")
	viper.BindPFlag("vcenterurl", cmd.Flags().Lookup("vcenterurl"))

	return cmd
}

func vCenterDriverCmd(out io.Writer, errOut io.Writer) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		consumer, err := vcenter.NewConsumer(viper.GetString("vcenterurl"), true)
		if err != nil {
			return err
		}
		driver, err := makeDriver(consumer)
		if err != nil {
			return err
		}
		defer driver.Close()

		return driver.Run()
	}
}
