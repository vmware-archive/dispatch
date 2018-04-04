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
	"path"

	"github.com/go-openapi/spec"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

var (
	createFunctionLong = i18n.T(`Create dispatch function.`)
	// TODO: add examples
	createFunctionExample = i18n.T(``)
	schemaInFile          = ""
	schemaOutFile         = ""
	fnSecrets             = []string{}
	fnServices            = []string{}
)

// NewCmdCreateFunction creates command responsible for dispatch function creation.
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
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVar(&schemaInFile, "schema-in", "", "path to file with input validation schema")
	cmd.Flags().StringVar(&schemaOutFile, "schema-out", "", "path to file with output validation schema")
	cmd.Flags().StringArrayVar(&fnSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")
	cmd.Flags().StringArrayVar(&fnServices, "service", []string{}, "Service instances this function uses, can be specified multiple times or a comma-delimited string")
	return cmd
}

type cliFunction struct {
	models.Function
	FunctionPath  string `json:"functionPath"`
	SchemaInPath  string `json:"schemaInPath"`
	SchemaOutPath string `json:"schemaOutPath"`
}

// CallCreateFunction makes the API call to create a function
func CallCreateFunction(f interface{}) error {
	client := functionManagerClient()
	function := f.(*models.Function)

	params := &fnstore.AddFunctionParams{
		Body:    function,
		Context: context.Background(),
	}

	created, err := client.Store.AddFunction(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*function = *created.Payload
	return nil
}

func createFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	functionPath := args[2]
	function := &models.Function{
		Image:    &args[0],
		Name:     &args[1],
		Secrets:  fnSecrets,
		Services: fnServices,
		Tags:     []*models.Tag{},
	}
	if cmdFlagApplication != "" {
		function.Tags = append(function.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	codeFileContent, err := ioutil.ReadFile(functionPath)
	if err != nil {
		message := fmt.Sprintf("Error when reading content of %s", functionPath)
		return formatCliError(err, message)
	}
	codeEncoded := string(codeFileContent)
	function.Code = &codeEncoded

	var schemaIn, schemaOut *spec.Schema
	if schemaInFile != "" {
		fullPath := path.Join(workDir, schemaInFile)
		schemaInContent, err := ioutil.ReadFile(fullPath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", fullPath)
			return formatCliError(err, message)
		}
		schemaIn = new(spec.Schema)
		if err := json.Unmarshal(schemaInContent, schemaIn); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s", fullPath)
			return formatCliError(err, message)
		}
	}
	if schemaOutFile != "" {
		fullPath := path.Join(workDir, schemaOutFile)
		schemaOutContent, err := ioutil.ReadFile(fullPath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", fullPath)
			return formatCliError(err, message)
		}
		schemaOut = new(spec.Schema)
		if err := json.Unmarshal(schemaOutContent, schemaOut); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s", fullPath)
			return formatCliError(err, message)
		}
	}

	function.Schema = &models.Schema{
		In:  schemaIn,
		Out: schemaOut,
	}

	err = CallCreateFunction(function)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(function)
	}
	fmt.Fprintf(out, "Created function: %s\n", *function.Name)
	return nil
}
