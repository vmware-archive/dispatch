///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	cmdConfig.Contexts[cmdConfig.Current] = &dispatchConfig
	vsConfigJSON, err := json.MarshalIndent(cmdConfig, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshalling json")
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), vsConfigJSON, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration to file: %s", viper.ConfigFileUsed())
	}
	fmt.Printf("You have successfully logged out\n")
	return nil
}
