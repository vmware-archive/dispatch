///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

var (
	updateSecretLong = i18n.T(`Update a dispatch secret.
		SECRETS_FILE - the path to a .json file describing a secret

		Example: secret.json:
		{
			"secret-key": "secret-value"
		}`)

	// TODO: add examples
	updateSecretExample = i18n.T(`update a secret`)
)

func CallUpdateSecret(input interface{}) error {
	client := secretStoreClient()
	secretBody := input.(*models.Secret)

	params := secret.NewUpdateSecretParams()
	params.Secret = secretBody
	params.SecretName = *secretBody.Name
	params.Tags = []string{}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	_, err := client.Secret.UpdateSecret(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return err
}

func NewCmdUpdateSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret SECRETS_FILE",
		Short:   i18n.T("Update secret"),
		Long:    updateSecretLong,
		Example: createSecretExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateSecret(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func updateSecret(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	secretPath := args[0]
	var secret = models.Secret{}

	if secretPath != "" {
		secretContent, err := ioutil.ReadFile(secretPath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", secretPath)
			return formatCliError(err, message)
		}
		if err := json.Unmarshal(secretContent, &secret); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s with error %s", secretContent, err)
			return formatCliError(err, message)
		}
	}

	err := CallUpdateSecret(&secret)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated secret: %s\n", *secret.Name)
	return nil
}
