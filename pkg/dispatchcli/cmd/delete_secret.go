///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	secret "github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	models "github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

var (
	deleteSecretsLong = i18n.T(`Delete secrets.`)

	// TODO: add examples
	deleteSecretsExample = i18n.T(``)
)

// NewCmdDeleteSecret creates command responsible for deleting  secrets.
func NewCmdDeleteSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret SECRET_NAME",
		Short:   i18n.T("Delete secrets"),
		Long:    deleteSecretsLong,
		Example: deleteSecretsExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"secrets"},
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteSecret(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteSecret makes the API call to delete a secret
func CallDeleteSecret(s interface{}) error {
	client := secretStoreClient()
	secretModel := s.(*cliSecret)
	params := &secret.DeleteSecretParams{
		SecretName: *secretModel.Name,
		Context:    context.Background(),
		Tags:       []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	_, err := client.Secret.DeleteSecret(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	// No content is returned from secret... should return secret payload
	// like all other endpoints.
	// *secretModel = *deleted.Payload
	return nil
}

func deleteSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	secretModel := cliSecret{
		Secret: models.Secret{
			Name: &args[0],
		},
	}
	err := CallDeleteSecret(&secretModel)
	if err != nil {
		return err
	}
	return formatDeleteSecretOutput(out, false, []*models.Secret{&secretModel.Secret})
}

func formatDeleteSecretOutput(out io.Writer, list bool, secrets []*models.Secret) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(secrets)
		}
		return encoder.Encode(secrets[0])
	}
	for _, i := range secrets {
		_, err := fmt.Fprintf(out, "Deleted secret: %s\n", *i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
