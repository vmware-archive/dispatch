///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	functionModels "github.com/vmware/dispatch/pkg/function-manager/gen/models"
	imageModels "github.com/vmware/dispatch/pkg/image-manager/gen/models"
	secretModels "github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

var (
	createLong = i18n.T(`Create a resource. See subcommands for resources that can be created.`)

	// TODO: Add examples
	createExample = i18n.T(``)
	file          = i18n.T(``)
)

type modelAction func(interface{}) error

type importFunction struct {
	functionModels.Function
}

func importFile(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string, actionMap map[string]modelAction) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "Error reading file %s", file)
	}

	// Manually split up the yaml doc.  This is NOT a streaming parser.
	docs := bytes.Split(b, []byte("---"))

	type kind struct {
		Kind string `json:"kind"`
	}

	type output struct {
		BaseImages []*imageModels.BaseImage   `json:"baseImages"`
		Images     []*imageModels.Image       `json:"images"`
		Functions  []*functionModels.Function `json:"functions"`
		Secrets    []*secretModels.Secret     `json:"secrets"`
	}

	o := output{}

	for _, doc := range docs {
		k := kind{}
		err = yaml.Unmarshal(doc, &k)
		if err != nil {
			return errors.Wrapf(err, "Error decoding document %s", string(doc))
		}
		switch docKind := k.Kind; docKind {
		case "base-image":
			m := &imageModels.BaseImage{}
			err = yaml.Unmarshal(doc, &m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding base image document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.BaseImages = append(o.BaseImages, m)
		case "image":
			m := &imageModels.Image{}
			err = yaml.Unmarshal(doc, &m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding image document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Images = append(o.Images, m)
		case "function":
			m := &cliFunction{}
			err = yaml.Unmarshal(doc, &m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding function document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Functions = append(o.Functions, &m.Function)
		case "secret":
			m := &cliSecret{}
			err = yaml.Unmarshal(doc, &m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding secret document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Secrets = append(o.Secrets, &m.Secret)
		default:
			continue
		}
	}
	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(o)
	}
	return nil
}

// NewCmdCreate creates a command object for the "create" action.
// Currently, one must use subcommands for specific resources to be created.
// In future create should accept file or stdin with multiple resources specifications.
// TODO: add create command implementation
func NewCmdCreate(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   i18n.T("Create resources."),
		Long:    createLong,
		Example: createExample,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				runHelp(cmd, args)
				return
			}

			createMap := map[string]modelAction{
				"image":      CallCreateImage,
				"base-image": CallCreateBaseImage,
				"function":   CallCreateFunction,
				"secret":     CallCreateSecret,
			}

			err := importFile(out, errOut, cmd, args, createMap)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")

	cmd.AddCommand(NewCmdCreateBaseImage(out, errOut))
	cmd.AddCommand(NewCmdCreateImage(out, errOut))
	cmd.AddCommand(NewCmdCreateFunction(out, errOut))
	cmd.AddCommand(NewCmdCreateSecret(out, errOut))
	cmd.AddCommand(NewCmdCreateAPI(out, errOut))
	cmd.AddCommand(NewCmdCreateSubscription(out, errOut))
	cmd.AddCommand(NewCmdCreateEventDriver(out, errOut))
	cmd.AddCommand(NewCmdCreateApplication(out, errOut))
	return cmd
}
