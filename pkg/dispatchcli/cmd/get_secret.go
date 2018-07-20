///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"golang.org/x/net/context"
)

var (
	getSecretsLong = i18n.T(`Get secrets.`)

	// TODO: add examples
	getSecretsExample = i18n.T(``)

	getSecretContent = false
)

// NewCmdGetSecret creates command responsible for getting secrets.
func NewCmdGetSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret [SECRET_NAME ...]",
		Short:   i18n.T("Get secrets"),
		Long:    getSecretsLong,
		Example: getSecretsExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"secrets"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := secretStoreClient()
			if len(args) == 1 {
				err = getSecret(out, errOut, cmd, args, c)
			} else {
				err = getSecrets(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().BoolVarP(&getSecretContent, "all", "", false, "also get secret content (in json format)")
	return cmd
}

func getSecret(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.SecretsClient) error {
	secretName := args[0]

	resp, err := c.GetSecret(context.TODO(), dispatchConfig.Organization, secretName)
	if err != nil {
		return err
	}

	return formatSecretOutput(out, false, []v1.Secret{*resp})
}

func getSecrets(out, errOut io.Writer, cmd *cobra.Command, c client.SecretsClient) error {
	resp, err := c.ListSecrets(context.TODO(), dispatchConfig.Organization)
	if err != nil {
		return err
	}
	return formatSecretOutput(out, true, resp)
}

func formatSecretOutput(out io.Writer, list bool, secrets []v1.Secret) error {

	if getSecretContent {
		dispatchConfig.Output = "json"
	}
	if w, err := formatOutput(out, list, secrets); w {
		return err
	}

	fmt.Fprintf(out, "Note: secret values are hidden, please use --all flag to get them\n\n")

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"ID", "Name", "Content"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, secret := range secrets {
		table.Append([]string{secret.ID.String(), *secret.Name, "<hidden>"})
	}
	table.Render()
	return nil
}
