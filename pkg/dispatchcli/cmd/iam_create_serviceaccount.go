///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"

	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"
)

var (
	createServiceAccountLong = i18n.T(`Create a dispatch service account`)

	createServiceAccountExample = i18n.T(`
# Create a service account by specifying a public key (public key file path)
dispatch iam create serviceaccount test_service_account --public-key ./app_rsa.pub
`)

	publicKeyPath string
)

// NewCmdIamCreateServiceAccount creates command responsible for service account creation
func NewCmdIamCreateServiceAccount(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T(`serviceaccount SERVICE_ACCOUNT_NAME --public-key PUBLIC_KEY_PATH`),
		Short:   i18n.T(`Create service account`),
		Long:    createServiceAccountLong,
		Example: createServiceAccountExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := createServiceAccount(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&publicKeyPath, "public-key", "", "public key file path")
	return cmd
}

// CallCreateServiceAccount makes the api call to create a service account
func CallCreateServiceAccount(p interface{}) error {
	client := identityManagerClient()
	serviceAccountModel := p.(*models.ServiceAccount)

	params := &serviceaccount.AddServiceAccountParams{
		Body:    serviceAccountModel,
		Context: context.Background(),
	}

	created, err := client.Serviceaccount.AddServiceAccount(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*serviceAccountModel = *created.Payload
	return nil
}

func createServiceAccount(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	serviceAccountName := args[0]
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("Error reading public key file: %s", err.Error())
	}
	publicKey := string(base64.StdEncoding.EncodeToString(publicKeyBytes))
	serviceAccountModel := &models.ServiceAccount{
		Name:      &serviceAccountName,
		PublicKey: &publicKey,
	}

	err = CallCreateServiceAccount(serviceAccountModel)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Create service account: %s\n", *serviceAccountModel.Name)
	return nil
}
