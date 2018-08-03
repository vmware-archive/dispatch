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
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	manageContextLong    = i18n.T(`Manage configuration context.`)
	manageContextExample = i18n.T(`
# list all contexts:
$ dispatch manage context`)

	currentContextLong = i18n.T(`Show current configuration context.`)

	setContextLong    = i18n.T(`Set configuration context.`)
	setContextExample = i18n.T(`
# set current context to localhost:
$ dispatch manage context set localhost
Set context to localhost`)

	deleteContextLong    = i18n.T(`Delete configuration context.`)
	deleteContextExample = i18n.T(`
# delete context localhost:
$ dispatch manage context delete localhost
Deleted context localhost`)
)

// NewCmdManageContext handles configuration context operations
func NewCmdManageContext(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Short:   i18n.T("Manage context"),
		Long:    manageContextLong,
		Example: manageContextExample,
		Aliases: []string{"app"},
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return formatContextOutput(out)
		},
	}

	cmd.AddCommand(NewCmdManageContextCurrent(out, errOut))
	cmd.AddCommand(NewCmdManageContextSet(out, errOut))
	cmd.AddCommand(NewCmdManageContextDelete(out, errOut))

	return cmd
}

func setContext(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	currentContext := args[0]
	if _, ok := cmdConfig.Contexts[currentContext]; !ok {
		return errors.Errorf("No such context %s", currentContext)
	}
	cmdConfig.Current = currentContext
	b, _ := json.MarshalIndent(cmdConfig, "", "    ")
	path := viper.ConfigFileUsed()
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return errors.Errorf("Failed to write config file %s", path)
	}
	fmt.Fprintf(out, "Set context to %s\n", currentContext)
	return nil
}

func deleteContext(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	ctxName := args[0]
	if _, ok := cmdConfig.Contexts[ctxName]; !ok {
		return errors.Errorf("No such context %s", ctxName)
	}

	delete(cmdConfig.Contexts, ctxName)
	b, _ := json.MarshalIndent(cmdConfig, "", "    ")
	path := viper.ConfigFileUsed()
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return errors.Errorf("Failed to write config file %s", path)
	}
	fmt.Fprintf(out, "Deleted context %s\n", ctxName)
	return nil
}

func formatContextOutput(out io.Writer) error {
	if w, err := formatOutput(out, false, cmdConfig); w {
		return err
	}

	headers := []string{"Context", "Config"}
	table := tablewriter.NewWriter(out)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	table.SetAutoWrapText(false)
	for context, config := range cmdConfig.Contexts {
		// Remove the cookie from output
		config.Cookie = ""
		configContent, _ := json.MarshalIndent(config, "", "  ")
		if context == cmdConfig.Current {
			context = fmt.Sprintf("* %s", context)
		}
		row := []string{context, string(configContent)}
		table.Append(row)
	}
	table.Render()
	return nil
}

func formatCurrentContextOutput(out io.Writer) error {
	if w, err := formatOutput(out, false, cmdConfig); w {
		return err
	}

	context := cmdConfig.Current
	if config, ok := cmdConfig.Contexts[context]; ok {
		headers := []string{"Context", "Config"}
		table := tablewriter.NewWriter(out)
		table.SetHeader(headers)
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetCenterSeparator("")
		table.SetAutoWrapText(false)
		// Remove the cookie from output
		config.Cookie = ""
		configContent, _ := json.MarshalIndent(config, "", "  ")
		context = fmt.Sprintf("* %s", context)
		row := []string{context, string(configContent)}
		table.Append(row)
		table.Render()
	} else {
		fmt.Fprintf(out, "No such current context was found %s\n", context)
	}
	return nil
}

func formatContextName(name string) string {
	return strings.ToLower(strings.Replace(name, ".", "-", -1))
}

// NewCmdManageContextCurrent show current context
func NewCmdManageContextCurrent(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: i18n.T("Show current context"),
		Long:  currentContextLong,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return formatCurrentContextOutput(out)
		},
	}
	return cmd
}

// NewCmdManageContextSet handles set context operations
func NewCmdManageContextSet(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set CONTEXT",
		Short:   i18n.T("Set context"),
		Long:    setContextLong,
		Example: setContextExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := setContext(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

// NewCmdManageContextDelete handles delete context operations
func NewCmdManageContextDelete(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete CONTEXT",
		Short:   i18n.T("Delete context"),
		Long:    deleteContextLong,
		Example: deleteContextExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteContext(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}
