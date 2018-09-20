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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/utils"
)

var (
	createFunctionLong = i18n.T(`Create dispatch function.`)
	// TODO: add examples
	createFunctionExample = i18n.T(``)
	functionImage         = ""
	depsImage             = ""
	handler               = ""
	schemaInFile          = ""
	schemaOutFile         = ""
	fnSecrets             []string
	fnServices            []string
	timeout               int64
)

// NewCmdCreateFunction creates command responsible for dispatch function creation.
func NewCmdCreateFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "function NAME PATH --image=IMAGE --handler=HANDLER [--schema-in=FILE] [--schema-out=FILE]",
		// Use:     "function NAME --function-image=IMAGE",
		Short:   i18n.T("Create function"),
		Long:    createFunctionLong,
		Example: createFunctionExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := functionManagerClient()
			err := createFunction(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&depsImage, "image", "", "REQUIRED: image to build function on")
	cmd.Flags().StringVar(&handler, "handler", "", "REQUIRED: fully-qualified function impl name (e.g. Java class or Python def)")
	// cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	// cmd.Flags().StringVar(&functionImage, "function-image", "", "REQUIRED: image to build function on")
	cmd.Flags().StringVar(&schemaInFile, "schema-in", "", "path to file with input validation schema")
	cmd.Flags().StringVar(&schemaOutFile, "schema-out", "", "path to file with output validation schema")
	cmd.Flags().StringArrayVar(&fnSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")
	cmd.Flags().StringArrayVar(&fnServices, "service", []string{}, "Service instances this function uses, can be specified multiple times or a comma-delimited string")
	cmd.Flags().Int64Var(&timeout, "timeout", 0, "A timeout to limit function execution time (in milliseconds). Default: 0 (no timeout)")
	cmd.MarkFlagRequired("image")
	return cmd
}

// CallCreateFunction makes the API call to create a function
func CallCreateFunction(c client.FunctionsClient) ModelAction {
	return func(f interface{}) error {
		function := f.(*v1.Function)

		// How will we verify that the user has permissions for the org?
		created, err := c.CreateFunction(context.TODO(), "", function)
		if err != nil {
			return err
		}
		*function = *created
		return nil
	}
}

func createFunction(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.FunctionsClient) error {
	sourcePath := args[1]
	isDir, err := utils.IsDir(sourcePath)
	if err != nil {
		return err
	}
	if isDir && handler == "" {
		return fmt.Errorf("error creating function %s: handler is required, source path %s is a directory", args[0], sourcePath)
	}
	codeFileContent, err := utils.TarGzBytes(sourcePath)
	if err != nil {
		return errors.Wrapf(err, "error reading %s", sourcePath)
	}

	ioutil.WriteFile("tmp/test.tgz", codeFileContent, 0600)

	function := &v1.Function{
		Meta: v1.Meta{
			Name: args[0],
			Tags: []*v1.Tag{},
		},
		// FunctionImageURL: functionImage,
		Services: fnServices,
		Image:    depsImage,
		Source:   codeFileContent,
		Handler:  handler,
		Secrets:  fnSecrets,
		Timeout:  timeout,
	}

	var schemaIn, schemaOut *spec.Schema
	if schemaInFile != "" {
		fullPath := path.Join(workDir, schemaInFile)
		schemaInContent, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return errors.Wrapf(err, "error when reading content of %s", fullPath)
		}
		schemaIn = new(spec.Schema)
		if err := json.Unmarshal(schemaInContent, schemaIn); err != nil {
			return errors.Wrapf(err, "error when parsing JSON from %s", fullPath)
		}
	}
	if schemaOutFile != "" {
		fullPath := path.Join(workDir, schemaOutFile)
		schemaOutContent, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return errors.Wrapf(err, "error when reading content of %s", fullPath)
		}
		schemaOut = new(spec.Schema)
		if err := json.Unmarshal(schemaOutContent, schemaOut); err != nil {
			return errors.Wrapf(err, "error when parsing JSON from %s", fullPath)
		}
	}

	function.Schema = &v1.Schema{
		In:  schemaIn,
		Out: schemaOut,
	}

	if err := CallCreateFunction(c)(function); err != nil {
		return err
	}
	if w, err := formatOutput(out, false, function); w {
		return err
	}
	fmt.Fprintf(out, "Created function: %s\n", function.Meta.Name)
	return nil
}
