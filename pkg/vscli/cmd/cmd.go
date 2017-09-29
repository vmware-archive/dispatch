///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"
	"io"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var vsConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Organization string `json:"organization"`
	Cookie       string `json:"cookie"`
}

var validResources = i18n.T(`Valid resource types include:
	* functions
	* images
	* base-images
	* secrets
    `)

var vsConfigPath = ""

func initConfig() {
	// Don't forget to read config either from vsConfigPath or from home directory!
	if vsConfigPath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(vsConfigPath)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vs" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vs")
	}

	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 8080)
	viper.SetDefault("organization", "serverless")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	viper.Unmarshal(&vsConfig)
}

// NewVSCLI creates cobra object for top-level VMware serverless CLI
func NewVSCLI(in io.Reader, out, errOut io.Writer) *cobra.Command {
	// Parent command to which all subcommands are added.
	cobra.OnInitialize(initConfig)
	cmds := &cobra.Command{
		Use:   "vs",
		Short: i18n.T("vs allows to interact with VMware Serverless platform."),
		Long:  i18n.T("vs allows to interact with VMware Serverless platform."),
		Run:   runHelp,
	}
	cmds.PersistentFlags().StringVar(&vsConfigPath, "config", "", "config file (default is $HOME/.vs)")
	cmds.PersistentFlags().String("host", "127.0.0.1", "VMware Serverless host to connect to")
	cmds.PersistentFlags().Int("port", 8000, "Port which VMware Serverless is listening on")
	cmds.PersistentFlags().String("organization", "serverless", "Organization name")
	viper.BindPFlag("host", cmds.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", cmds.PersistentFlags().Lookup("port"))
	viper.BindPFlag("organization", cmds.PersistentFlags().Lookup("organization"))

	cmds.AddCommand(NewCmdGet(out, errOut))
	cmds.AddCommand(NewCmdCreate(out, errOut))
	cmds.AddCommand(NewCmdExec(out, errOut))
	cmds.AddCommand(NewCmdLogin(in, out, errOut))
	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
