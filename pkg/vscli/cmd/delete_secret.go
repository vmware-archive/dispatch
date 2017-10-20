///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	secret "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client/secret"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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
	return cmd
}

func deleteSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := secretStoreClient()
	params := &secret.DeleteSecretParams{
		Context:    context.Background(),
		SecretName: args[0],
	}
	_, err := client.Secret.DeleteSecret(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatDeleteSecretOutput(out, false, []*models.Secret{})
}

func formatDeleteSecretOutput(out io.Writer, list bool, secrets []*models.Secret) error {
	if vsConfig.Json {
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
