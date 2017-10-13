///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	secret "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client/secret"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
	"golang.org/x/net/context"
)

var (
	getSecretsLong = i18n.T(`Get secrets.`)

	// TODO: add examples
	getSecretsExample = i18n.T(``)
)

// NewCmdGetSecret creates command responsible for getting secrets.
func NewCmdGetSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret [SECRET_NAME]",
		Short:   i18n.T("Get secrets"),
		Long:    getSecretsLong,
		Example: getSecretsExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"secrets"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) > 0 {
				err = getSecret(out, errOut, cmd, args)
			} else {
				err = getSecrets(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := secretStoreClient()
	params := &secret.GetSecretParams{
		Context:    context.Background(),
		SecretName: args[0],
	}

	resp, err := client.Secret.GetSecret(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatSecretOutput(out, false, []*models.Secret{resp.Payload})
}

func getSecrets(out, errOut io.Writer, cmd *cobra.Command) error {
	client := secretStoreClient()
	params := &secret.GetSecretsParams{
		Context: context.Background(),
	}
	resp, err := client.Secret.GetSecrets(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	fmt.Printf("response: %v\n", resp.Payload)
	return formatSecretOutput(out, true, resp.Payload)
}

func formatSecretOutput(out io.Writer, list bool, secrets []*models.Secret) error {

	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(secrets)
		}
		return encoder.Encode(secrets[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"ID", "Name"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, secret := range secrets {
		table.Append([]string{secret.ID.String(), *secret.Name})
	}
	table.Render()
	return nil
}
