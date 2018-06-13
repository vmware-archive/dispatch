///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/version"
)

type versions struct {
	Client *v1.Version `json:"client,omitempty"`
	Server *v1.Version `json:"server,omitempty"`
}

// NewCmdVersion creates a version command for CLI
func NewCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: i18n.T("Print Dispatch CLI version."),
		Run: func(cmd *cobra.Command, args []string) {
			clientVersion := version.Get()
			if dispatchConfig.JSON {
				encoder := json.NewEncoder(out)
				encoder.SetIndent("", "  ")

				serverVersion, err := getServerVersion()
				CheckErr(err)

				encoder.Encode(&versions{Client: clientVersion, Server: serverVersion})
				return
			}

			fmt.Fprintf(out, "Client: %s\n", clientVersion.Version)
			serverVersion, err := getServerVersion()
			CheckErr(err)
			fmt.Fprintf(out, "Server: %s\n", serverVersion.Version)
		},
	}
	return cmd
}

func getServerVersion() (*v1.Version, error) {
	v, err := identityManagerClient().GetVersion(context.Background())
	if err != nil {
		return nil, err
	}
	return v, nil
}
