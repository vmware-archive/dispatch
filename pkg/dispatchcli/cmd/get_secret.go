///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	secret "github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	models "github.com/vmware/dispatch/pkg/secret-store/gen/models"
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
			if len(args) == 1 {
				err = getSecret(out, errOut, cmd, args)
			} else {
				err = getSecrets(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().BoolVarP(&getSecretContent, "all", "", false, "also get secret content (in json format)")
	return cmd
}

func getSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := secretStoreClient()
	params := &secret.GetSecretParams{
		Context:    context.Background(),
		SecretName: args[0],
		Tags:       []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Secret.GetSecret(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	if resp.Payload.Name == nil {
		err := secret.NewGetSecretNotFound()
		err.Payload = &models.Error{
			Code:    404,
			Message: &args[0],
		}
		return formatAPIError(err, params)
	}

	return formatSecretOutput(out, false, []*models.Secret{resp.Payload})
}

func getSecrets(out, errOut io.Writer, cmd *cobra.Command) error {
	client := secretStoreClient()
	params := &secret.GetSecretsParams{
		Context: context.Background(),
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Secret.GetSecrets(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatSecretOutput(out, true, resp.Payload)
}

func formatSecretOutput(out io.Writer, list bool, secrets []*models.Secret) error {

	if getSecretContent || dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(secrets)
		}
		return encoder.Encode(secrets[0])
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
