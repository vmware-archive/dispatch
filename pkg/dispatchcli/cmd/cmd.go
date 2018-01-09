///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var dispatchConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Organization string `json:"organization"`
	Cookie       string `json:"cookie"`
	SkipAuth     bool   `json:"skipauth"`
	Insecure     bool   `json:"insecure"`
	Json         bool   `json:"-"`
}

var validResources = i18n.T(`Valid resource types include:
	* api
	* base-images
	* event-driver
	* functions
	* images
	* secrets
	* subscription
    `)

var dispatchConfigPath = ""

func initConfig() {
	// Don't forget to read config either from dispatchConfigPath or from home directory!
	if dispatchConfigPath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(dispatchConfigPath)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".dispatch" (without extension).
		viper.AddConfigPath(path.Join(home, ".dispatch"))
		viper.SetConfigName("config")
	}
	// TODO (bjung): add config command to print config used
	viper.ReadInConfig()
	viper.Unmarshal(&dispatchConfig)
}

// NewCLI creates cobra object for top-level Dispatch CLI
func NewCLI(in io.Reader, out, errOut io.Writer) *cobra.Command {
	// Parent command to which all subcommands are added.
	cobra.OnInitialize(initConfig)
	cmds := &cobra.Command{
		Use:   "dispatch",
		Short: i18n.T("dispatch allows to interact with VMware Dispatch framework."),
		Long:  i18n.T("dispatch allows to interact with VMware Dispatch framework."),
		Run:   runHelp,
	}
	cmds.PersistentFlags().StringVar(&dispatchConfigPath, "config", "", "config file (default is $HOME/.dispatch)")
	cmds.PersistentFlags().String("host", "dispatch.vmware.com", "VMware Dispatch host to connect to")
	cmds.PersistentFlags().Int("port", 443, "Port which VMware Dispatch is listening on")
	cmds.PersistentFlags().String("organization", "dispatch", "Organization name")
	cmds.PersistentFlags().Bool("insecure", false, "If true, will ignore verifying the server's certificate and your https connection is insecure.")
	cmds.PersistentFlags().Bool("skipauth", false, "skip authentication (only take effect with a SkipAuthMode-enabled server)")
	cmds.PersistentFlags().BoolVar(&dispatchConfig.Json, "json", false, "Output raw JSON")
	viper.BindPFlag("host", cmds.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", cmds.PersistentFlags().Lookup("port"))
	viper.BindPFlag("organization", cmds.PersistentFlags().Lookup("organization"))
	viper.BindPFlag("skipauth", cmds.PersistentFlags().Lookup("skipauth"))
	viper.BindPFlag("insecure", cmds.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("json", cmds.PersistentFlags().Lookup("json"))

	cmds.AddCommand(NewCmdGet(out, errOut))
	cmds.AddCommand(NewCmdCreate(out, errOut))
	cmds.AddCommand(NewCmdExec(out, errOut))
	cmds.AddCommand(NewCmdDelete(out, errOut))
	cmds.AddCommand(NewCmdLogin(in, out, errOut))
	cmds.AddCommand(NewCmdLogout(in, out, errOut))
	cmds.AddCommand(NewCmdEmit(out, errOut))
	cmds.AddCommand(NewCmdInstall(out, errOut))
	cmds.AddCommand(NewCmdUninstall(out, errOut))
	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
