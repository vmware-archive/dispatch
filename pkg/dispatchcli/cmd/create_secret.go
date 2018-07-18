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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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

// NewCmdCreateSecret creates command responsible for secret creation.
func NewCmdCreateSecret(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret SECRET_NAME SECRETS_FILE",
		Short:   i18n.T("Create secret"),
		Long:    createSecretLong,
		Example: createSecretExample,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := secretStoreClient()
			err := createSecret(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	return cmd
}

// CallCreateSecret makes the API call to create a secret
func CallCreateSecret(c client.SecretsClient) ModelAction {
	return func(s interface{}) error {
		secretModel := s.(*v1.Secret)

		created, err := c.CreateSecret(context.TODO(), dispatchConfig.Organization, secretModel)
		if err != nil {
			return err
		}

		*secretModel = *created
		return nil
	}
}

func createSecret(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.SecretsClient) error {
	secretPath := args[1]

	body := &v1.Secret{
		Name: &args[0],
	}

	if secretPath != "" {
		secretContent, err := ioutil.ReadFile(secretPath)
		if err != nil {
			return errors.Wrapf(err, "error when reading content of %s", secretPath)
		}
		if err := json.Unmarshal(secretContent, &body.Secrets); err != nil {
			return errors.Wrapf(err, "Error when parsing JSON from %s", secretContent)
		}
	}

	if cmdFlagApplication != "" {
		body.Tags = append(body.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}
	err := CallCreateSecret(c)(body)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, body); w {
		return err
	}
	fmt.Fprintf(out, "Created secret: %s\n", *body.Name)
	return nil
}
