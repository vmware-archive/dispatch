///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	loginLong = i18n.T(`Login to VMware serverless platform.`)

	// TODO: Add examples
	loginExample = i18n.T(``)
)

// NewCmdLogin creates a command to login to VMware serverless platform.
func NewCmdLogin(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		Short:   i18n.T("Login to VMware serverless platform."),
		Long:    loginLong,
		Example: loginExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := login(in, out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func login(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
