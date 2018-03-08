///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/version"
)

// NewCmdVersion creates a version command for CLI
func NewCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: i18n.T("Print Dispatch CLI version."),
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			fmt.Fprintf(out, "%s-%s\n", v.Version, v.Commit[0:7])
		},
	}
	return cmd
}
