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

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	secret "github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	models "github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

var (
	createSecretLong = i18n.T(`Create a dispatch secret.
	SECRETS_FILE - the path to a .json file contains secrets

	Example: secret.json:
		{
			"secret-key": "secret-value"
		}`)

	// TODO: add examples
	createSecretExample = i18n.T(`create a secret`)
)

type cliSecret struct {
	models.Secret
	SecretPath string `json:"secretPath"`
}

// CallCreateSecret makes the API call to create a secret
func CallCreateSecret(s interface{}) error {
	client := secretStoreClient()
	body := s.(*cliSecret)

	if body.SecretPath != "" {
		secretContent, err := ioutil.ReadFile(body.SecretPath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", body.SecretPath)
			return formatCliError(err, message)
		}
		if err := json.Unmarshal(secretContent, &body.Secrets); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s", secretContent)
			return formatCliError(err, message)
		}
	}

	params := &secret.AddSecretParams{
		Secret:  &body.Secret,
		Context: context.Background(),
	}
	created, err := client.Secret.AddSecret(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	body.Secret = *created.Payload
	return nil
}

// NewCmdCreateSecret creates command responsible for secret creation.
func NewCmdCreateSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret SECRET_NAME SECRETS_FILE",
		Short:   i18n.T("Create secret"),
		Long:    createSecretLong,
		Example: createSecretExample,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createSecret(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	return cmd
}

func createSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	body := &cliSecret{
		Secret: models.Secret{
			Name: &args[0],
			Tags: models.SecretTags{},
		},
		SecretPath: args[1],
	}
	if cmdFlagApplication != "" {
		body.Secret.Tags = append(body.Secret.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}
	err := CallCreateSecret(body)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(body.Secret)
	}
	fmt.Fprintf(out, "Created secret: %s\n", *body.Name)
	return nil
}
