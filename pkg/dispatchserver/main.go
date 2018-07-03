///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var dispatchConfigPath = ""

// NewCLI creates cobra object for top-level Dispatch server
func NewCLI(out io.Writer) *cobra.Command {
	log.SetOutput(out)
	cmd := &cobra.Command{
		Use:   "dispatch-server",
		Short: i18n.T("Dispatch is a batteries-included serverless framework."),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig(cmd, defaultConfig)
		},
	}
	cmd.SetOutput(out)

	configGlobalFlags(cmd.PersistentFlags())

	cmd.AddCommand(NewCmdLocal(out, defaultConfig))
	cmd.AddCommand(NewCmdFunctions(out, defaultConfig))

	return cmd
}

func initConfig(cmd *cobra.Command, targetConfig *serverConfig) {
	v := viper.New()
	configPath := os.Getenv("DISPATCH_CONFIG")

	if dispatchConfigPath != "" {
		configPath = dispatchConfigPath
	}

	if configPath != "" {
		v.SetConfigFile(dispatchConfigPath)
		if err := v.ReadInConfig(); err != nil {
			log.Fatalf("Unable to read the config file: %s", err)
		}
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		v.BindPFlag(f.Name, f)
		v.BindEnv(f.Name, "DISPATCH_"+strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)))
	})
	err := v.Unmarshal(targetConfig)
	if err != nil {
		log.Fatalf("Unable to create configuration: %s", err)
	}
	if defaultConfig.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func bindLocalFlags(target interface{}) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			v := viper.New()
			v.BindPFlag(f.Name, f)
			v.BindEnv(f.Name, "DISPATCH_"+strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)))

			err := v.Unmarshal(target)
			if err != nil {
				log.Fatalf("Unable to create configuration: %s", err)
			}
		})
	}
}
