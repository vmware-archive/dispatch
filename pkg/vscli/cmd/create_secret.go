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
	createSecretLong = i18n.T(`Create serverless secret.`)

	// TODO: add examples
	createSecretExample = i18n.T(``)
)

// NewCmdCreateSecret creates command responsible for secret creation.
func NewCmdCreateSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret [--name SECRET_NAME] SECRETS_FILE",
		Short:   i18n.T("Create secret"),
		Long:    createSecretLong,
		Example: createSecretExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := createSecret(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func createSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
