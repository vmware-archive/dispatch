///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	logoutLong = i18n.T(`Logout from VMware Dispatch.`)
	// TODO: Add examples
	logoutExample = i18n.T(``)
)

// NewCmdLogout creates a command to logout from Dispatch.
func NewCmdLogout(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logout",
		Short:   i18n.T("Logout from VMware Dispatch."),
		Long:    logoutLong,
		Example: logoutExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := logout(in, out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func logout(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	dispatchConfig.Cookie = ""
	dispatchConfig.ServiceAccount = ""
	dispatchConfig.JWTPrivateKey = ""
	writeConfigFile()
	fmt.Printf("You have successfully logged out\n")
	return nil
}
