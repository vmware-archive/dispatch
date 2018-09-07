///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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
			c := secretStoreClient()
			err := deleteSecret(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteSecret makes the API call to delete a secret
func CallDeleteSecret(c client.SecretsClient) ModelAction {
	return func(s interface{}) error {
		secretModel := s.(*v1.Secret)

		err := c.DeleteSecret(context.TODO(), dispatchConfig.Organization, secretModel.Meta.Name)
		if err != nil {
			return err
		}
		// No content is returned from secret... should return secret payload
		// like all other endpoints.
		// *secretModel = *deleted.Payload
		return nil
	}
}

func deleteSecret(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.SecretsClient) error {
	secretModel := v1.Secret{
		Meta: v1.Meta{
			Name: args[0],
		},
	}
	err := CallDeleteSecret(c)(&secretModel)
	if err != nil {
		return err
	}
	return formatDeleteSecretOutput(out, false, []*v1.Secret{&secretModel})
}

func formatDeleteSecretOutput(out io.Writer, list bool, secrets []*v1.Secret) error {
	if w, err := formatOutput(out, list, secrets); w {
		return err
	}
	for _, i := range secrets {
		_, err := fmt.Fprintf(out, "Deleted secret: %s\n", *i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
