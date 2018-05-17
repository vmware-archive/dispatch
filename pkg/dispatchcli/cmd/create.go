///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/utils"
)

var (
	createLong = i18n.T(`Create a resource. See subcommands for resources that can be created.`)

	// TODO: Add examples
	createExample = i18n.T(``)
	file          = i18n.T(``)
	workDir       = i18n.T(``)
)

// ModelAction is the function type for CLI actions
type ModelAction func(interface{}) error

type importFunction struct {
	v1.Function
}

func resolveFileReference(ref string) (string, error) {
	if strings.HasPrefix(ref, "@") {
		filePath := path.Join(workDir, (ref)[1:])
		fileContent, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", errors.Wrapf(err, "Error when reading content of %s", filePath)
		}
		encoded := string(fileContent)
		return encoded, nil
	}
	return ref, nil
}

func importFile(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string, actionMap map[string]ModelAction, actionName string) error {
	fullPath := path.Join(workDir, file)
	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return errors.Wrapf(err, "Error reading file %s", fullPath)
	}

	// Manually split up the yaml doc.  This is NOT a streaming parser.
	docs := bytes.Split(b, []byte("---"))

	type kind struct {
		Kind string `json:"kind"`
	}

	type output struct {
		APIs             []*v1.API             `json:"api"`
		BaseImages       []*v1.BaseImage       `json:"baseImages"`
		Images           []*v1.Image           `json:"images"`
		DriverTypes      []*v1.EventDriverType `json:"driverTypes"`
		Drivers          []*v1.EventDriver     `json:"drivers"`
		Subscriptions    []*v1.Subscription    `json:"subscriptions"`
		Functions        []*v1.Function        `json:"functions"`
		Secrets          []*v1.Secret          `json:"secrets"`
		Policies         []*v1.Policy          `json:"policies"`
		ServiceInstances []*v1.ServiceInstance `json:"serviceInstances"`
		ServiceAccounts  []*v1.ServiceAccount  `json:"serviceaccounts"`
	}

	o := output{}

	for _, doc := range docs {
		k := &kind{}
		err = yaml.Unmarshal(doc, k)
		if err != nil {
			return errors.Wrapf(err, "Error decoding document %s", string(doc))
		}
		switch docKind := k.Kind; docKind {
		case utils.APIKind:
			m := &v1.API{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding api document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.APIs = append(o.APIs, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.BaseImageKind:
			m := &v1.BaseImage{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding base image document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.BaseImages = append(o.BaseImages, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.ImageKind:
			m := &v1.Image{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding image document %s", string(doc))
			}
			if m.RuntimeDependencies != nil {
				manifest, err := resolveFileReference(m.RuntimeDependencies.Manifest)
				if err != nil {
					return err
				}
				m.RuntimeDependencies.Manifest = manifest
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Images = append(o.Images, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.FunctionKind:
			m := &v1.Function{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding function document %s", string(doc))
			}
			if m.SourcePath != "" {
				sourcePath := filepath.Join(workDir, m.SourcePath)
				isDir, err := utils.IsDir(sourcePath)
				if err != nil {
					return formatCliError(err, fmt.Sprintf("Error determining id source path is a dir: '%s'", sourcePath))
				}
				if isDir && m.Handler == "" {
					return formatCliError(errors.New("--handler is required"), "handler is required: source path is a directory")
				}
				sourceTarGz, err := utils.TarGzBytes(sourcePath)
				if err != nil {
					return errors.Wrapf(err, "Error when reading content of %s", sourcePath)
				}
				m.Source = sourceTarGz
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Functions = append(o.Functions, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.DriverTypeKind:
			m := &v1.EventDriverType{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error when decoding driver type document of %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.DriverTypes = append(o.DriverTypes, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.DriverKind:
			m := &v1.EventDriver{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding driver document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Drivers = append(o.Drivers, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.SubscriptionKind:
			m := &v1.Subscription{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding subscription document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Subscriptions = append(o.Subscriptions, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.SecretKind:
			m := &v1.Secret{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding secret document %s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Secrets = append(o.Secrets, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.PolicyKind:
			m := &v1.Policy{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding policy document &s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.Policies = append(o.Policies, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.ServiceInstanceKind:
			m := &v1.ServiceInstance{}
			err := yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding service instance document &s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.ServiceInstances = append(o.ServiceInstances, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		case utils.ServiceAccountKind:
			m := &v1.ServiceAccount{}
			err = yaml.Unmarshal(doc, m)
			if err != nil {
				return errors.Wrapf(err, "Error decoding service account document &s", string(doc))
			}
			err = actionMap[docKind](m)
			if err != nil {
				return err
			}
			o.ServiceAccounts = append(o.ServiceAccounts, m)
			fmt.Fprintf(out, "%s %s: %s\n", actionName, docKind, *m.Name)
		default:
			continue
		}
	}
	if dispatchConfig.JSON {
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

			fnClient := functionManagerClient()
			imgClient := imageManagerClient()
			eventClient := eventManagerClient()
			apiClient := apiManagerClient()

			createMap := map[string]ModelAction{
				utils.ImageKind:           CallCreateImage(imgClient),
				utils.BaseImageKind:       CallCreateBaseImage(imgClient),
				utils.FunctionKind:        CallCreateFunction(fnClient),
				utils.SecretKind:          CallCreateSecret,
				utils.ServiceInstanceKind: CallCreateServiceInstance,
				utils.PolicyKind:          CallCreatePolicy,
				utils.ApplicationKind:     CallCreateApplication,
				utils.ServiceAccountKind:  CallCreateServiceAccount,
				utils.DriverTypeKind:      CallCreateEventDriverType(eventClient),
				utils.DriverKind:          CallCreateEventDriver(eventClient),
				utils.SubscriptionKind:    CallCreateSubscription(eventClient),
				utils.APIKind:             CallCreateAPI(apiClient),
			}

			err := importFile(out, errOut, cmd, args, createMap, "Created")
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	cmd.AddCommand(NewCmdCreateBaseImage(out, errOut))
	cmd.AddCommand(NewCmdCreateImage(out, errOut))
	cmd.AddCommand(NewCmdCreateFunction(out, errOut))
	cmd.AddCommand(NewCmdCreateSecret(out, errOut))
	cmd.AddCommand(NewCmdCreateAPI(out, errOut))
	cmd.AddCommand(NewCmdCreateSubscription(out, errOut))
	cmd.AddCommand(NewCmdCreateEventDriver(out, errOut))
	cmd.AddCommand(NewCmdCreateEventDriverType(out, errOut))
	cmd.AddCommand(NewCmdCreateApplication(out, errOut))
	cmd.AddCommand(NewCmdCreateServiceInstance(out, errOut))
	return cmd
}
