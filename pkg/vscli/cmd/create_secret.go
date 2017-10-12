///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	secret "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client/secret"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	createSecretLong = i18n.T(`Create serverless secret.`)

	// TODO: add examples
	createSecretExample = i18n.T(``)
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
			err := createSecret(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func createSecret(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	var secretValue models.SecretValue
	secretFile := args[1]
	if secretFile != "" {
		secretContent, err := ioutil.ReadFile(secretFile)
		if err != nil {
			fmt.Fprintf(errOut, "Error when reading content of %s\n", secretFile)
			return err
		}
		// secretValue = new(models.SecretValue)
		if err := json.Unmarshal(secretContent, &secretValue); err != nil {
			fmt.Fprintf(errOut, "Error when parsing JSON from %s\n", schemaInFile)
			return err
		}
	}
	fmt.Printf("secret file %v\n", secretValue)

	client := secretStoreClient()
	body := &models.Secret{
		//TODO: ID should be generated from the server side
		// ID:      "id",
		Name:    &args[0],
		Secrets: secretValue,
	}
	params := &secret.AddSecretParams{
		Secret:  body,
		Context: context.Background(),
	}
	created, err := client.Secret.AddSecret(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Printf("created secret: %s\n", *created.Payload.Name)
	return nil
}
