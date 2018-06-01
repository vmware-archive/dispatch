///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/go-openapi/spec"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/utils"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
)

var (
	createFunctionLong = i18n.T(`Create dispatch function.`)
	// TODO: add examples
	createFunctionExample = i18n.T(``)
	depsImage             = ""
	handler               = ""
	schemaInFile          = ""
	schemaOutFile         = ""
	fnSecrets             = []string{}
	fnServices            = []string{}
	timeout               int64
)

// NewCmdCreateFunction creates command responsible for dispatch function creation.
func NewCmdCreateFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function NAME PATH --image=IMAGE --handler=HANDLER [--schema-in=FILE] [--schema-out=FILE]",
		Short:   i18n.T("Create function"),
		Long:    createFunctionLong,
		Example: createFunctionExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createFunction(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&depsImage, "image", "", "REQUIRED: image to build function on")
	cmd.Flags().StringVar(&handler, "handler", "", "REQUIRED: fully-qualified function impl name (e.g. Java class or Python def)")
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVar(&schemaInFile, "schema-in", "", "path to file with input validation schema")
	cmd.Flags().StringVar(&schemaOutFile, "schema-out", "", "path to file with output validation schema")
	cmd.Flags().StringArrayVar(&fnSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")
	cmd.Flags().StringArrayVar(&fnServices, "service", []string{}, "Service instances this function uses, can be specified multiple times or a comma-delimited string")
	cmd.Flags().Int64Var(&timeout, "timeout", 0, "A timeout to limit function execution time.")
	cmd.MarkFlagRequired("image")
	return cmd
}

type cliFunction struct {
	v1.Function
	FunctionPath  string `json:"functionPath"`
	SchemaInPath  string `json:"schemaInPath"`
	SchemaOutPath string `json:"schemaOutPath"`
}

// CallCreateFunction makes the API call to create a function
func CallCreateFunction(f interface{}) error {
	client := functionManagerClient()
	function := f.(*v1.Function)

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
	sourcePath := args[1]
	isDir, err := utils.IsDir(sourcePath)
	if err != nil {
		return formatCliError(err, fmt.Sprintf("Error determining id source path is a dir: '%s'", sourcePath))
	}
	if isDir && handler == "" {
		return formatCliError(errors.New("--handler is required"), "handler is required: source path is a directory")
	}
	codeFileContent, err := utils.TarGzBytes(sourcePath)
	if err != nil {
		message := fmt.Sprintf("Error reading %s", sourcePath)
		return formatCliError(err, message)
	}
	function := &v1.Function{
		Image:    &depsImage,
		Name:     &args[0],
		Source:   codeFileContent,
		Handler:  handler,
		Secrets:  fnSecrets,
		Services: fnServices,
		Timeout:  timeout,
		Tags:     []*v1.Tag{},
	}
	if cmdFlagApplication != "" {
		function.Tags = append(function.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

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

	function.Schema = &v1.Schema{
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
