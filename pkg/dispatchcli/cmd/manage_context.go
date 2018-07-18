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
	manageContextLong = i18n.T(`Manage configuration context.`)

	// TODO: add examples
	manageContextExample = i18n.T(``)
	currentContext       = i18n.T(``)
)

// NewCmdManageContext handles configuration context operations
func NewCmdManageContext(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context [--set CONTEXT]",
		Short:   i18n.T("Manage context"),
		Long:    manageContextLong,
		Example: manageContextExample,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"app"},
		Run: func(cmd *cobra.Command, args []string) {
			err := manageContext(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&currentContext, "set", "s", "", "Set current context")
	return cmd
}

func manageContext(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	if currentContext != "" {
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
	return formatContextOutput(out)
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

func formatContextName(name string) string {
	return strings.ToLower(strings.Replace(name, ".", "-", -1))
}
