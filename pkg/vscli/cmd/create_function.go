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

	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/store"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	createFunctionLong = i18n.T(`Create serverless function.`)
	// TODO: add examples
	createFunctionExample = i18n.T(``)
	schemaInFile          = ""
	schemaOutFile         = ""
)

// NewCmdCreateFunction creates command responsible for serverless function creation.
func NewCmdCreateFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function IMAGE_NAME FUNCTION_NAME FUNCTION_FILE [--schema-in SCHEMA_FILE] [--schema-out SCHEMA_FILE]",
		Short:   i18n.T("Create function"),
		Long:    createFunctionLong,
		Example: createFunctionExample,
		Args:    cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			err := createFunction(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&schemaInFile, "schema-in", "", "path to file with input validation schema")
	cmd.Flags().StringVar(&schemaOutFile, "schema-out", "", "path to file with output validation schema")
	return cmd
}

func createFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	codeFilePath := args[2]
	codeFileContent, err := ioutil.ReadFile(codeFilePath)
	if err != nil {
		fmt.Fprintf(errOut, "Error when reading content of %s\n", codeFilePath)
		return err
	}
	codeEncoded := string(codeFileContent)

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
		Image:  &args[0],
		Name:   &args[1],
		Schema: &schema,
	}

	params := &fnstore.AddFunctionParams{
		Body:    function,
		Context: context.Background(),
	}
	client := functionManagerClient()
	created, err := client.Store.AddFunction(params)
	if err != nil {
		fmt.Fprintf(errOut, "Error when creating a function %s\n", *function.Name)
		return err
	}
	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Fprintf(out, "Created function: %s\n", *created.Payload.Name)
	return nil
}
