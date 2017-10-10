///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
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

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	logoutLong = i18n.T(`Logout from VMware serverless platform.`)
	// TODO: Add examples
	logoutExample = i18n.T(``)
)

// NewCmdLogout creates a command to logout from VMware serverless platform.
func NewCmdLogout(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logout",
		Short:   i18n.T("Logout from VMware serverless platform."),
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

	vsConfig.Cookie = ""
	vsConfigJSON, err := json.MarshalIndent(vsConfig, "", "    ")
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
