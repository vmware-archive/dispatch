///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/client/store"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	createFunctionLong = i18n.T(`Create serverless function.`)
	// TODO: add examples
	createFunctionExample = i18n.T(``)
	functionName          = ""
	imageName             = ""
	schemaInFile          = ""
	schemaOutFile         = ""
)

// NewCmdCreateFunction creates command responsible for serverless function creation.
func NewCmdCreateFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function --name FUNCTION_NAME --image IMAGE_NAME [--schema-in SCHEMA_FILE] [--schema-out SCHEMA_FILE] FUNCTION_FILE",
		Short:   i18n.T("Create function"),
		Long:    createFunctionLong,
		Example: createFunctionExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := createFunction(out, errOut, cmd, args)
			CheckErr(err)
		},
		PreRunE: validateFnCreateFunc(errOut),
	}
	cmd.Flags().StringVar(&functionName, "name", "", "Function name. Will be used as its identifier (Required)")
	cmd.Flags().StringVar(&imageName, "image", "", "Image name to use for this function (Required)")
	cmd.Flags().StringVar(&schemaInFile, "schema-in", "", "path to file with input validation schema")
	cmd.Flags().StringVar(&schemaOutFile, "schema-out", "", "path to file with output validation schema")
	return cmd
}

func validateFnCreateFunc(errOut io.Writer) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if imageName == "" {
			fmt.Fprintf(errOut, "--image must be provided")
			return errors.New("missing required --image parameter")
		}
		if functionName == "" {
			fmt.Fprintf(errOut, "--name must be provided")
			return errors.New("missing required --name parameter")
		}
		return nil
	}
}

func createFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	codeFilePath := args[0]
	codeFileContent, err := ioutil.ReadFile(codeFilePath)
	if err != nil {
		fmt.Fprintf(errOut, "Error when reading content of %s\n", codeFilePath)
		return err
	}
	codeEncoded := base64.StdEncoding.EncodeToString(codeFileContent)

	var schemaInJSON, schemaOutJSON map[string]interface{}
	if schemaInFile != "" {
		schemaInContent, err := ioutil.ReadFile(schemaInFile)
		if err != nil {
			fmt.Fprintf(errOut, "Error when reading content of %s\n", schemaInFile)
			return err
		}

		err = json.Unmarshal(schemaInContent, &schemaInJSON)
		if err != nil {
			fmt.Fprintf(errOut, "Error when parsing JSON from %s\n", schemaInFile)
			return err
		}
	}
	if schemaOutFile != "" {
		schemaOutContent, err := ioutil.ReadFile(schemaOutFile)
		if err != nil {
			fmt.Fprintf(errOut, "Error when reading content of %s\n", schemaOutFile)
			return err
		}

		err = json.Unmarshal(schemaOutContent, &schemaOutJSON)
		if err != nil {
			fmt.Fprintf(errOut, "Error when parsing JSON from %s\n", schemaInFile)
			return err
		}

	}
	schema := models.Schema{
		In:  schemaInJSON,
		Out: schemaOutJSON,
	}

	function := &models.Function{
		Code:   &codeEncoded,
		Image:  &imageName,
		Name:   &functionName,
		Schema: &schema,
	}

	params := &fnstore.AddFunctionParams{
		Body:    function,
		Context: context.Background(),
	}
	client := functionManagerClient()
	created, err := client.Store.AddFunction(params)
	if err != nil {
		fmt.Fprintf(errOut, "Error when creating a function %s\n", functionName)
		return err
	}
	fmt.Fprintf(out, "Created function: %s\n", *created.Payload.Name)
	return nil
}
