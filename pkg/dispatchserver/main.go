///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"encoding/json"
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
	logFile, err := os.OpenFile("./dispatch.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	mw := io.MultiWriter(out, logFile)
	log.SetOutput(mw)

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
	cmd.AddCommand(NewCmdImages(out, defaultConfig))
	cmd.AddCommand(NewCmdSecrets(out, defaultConfig))
	cmd.AddCommand(NewCmdEvents(out, defaultConfig))
	cmd.AddCommand(NewCmdAPIs(out, defaultConfig))
	cmd.AddCommand(NewCmdIdentity(out, defaultConfig))
	cmd.AddCommand(NewCmdServices(out, defaultConfig))

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

func bindLocalFlags(targetStruct interface{}) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		v := viper.New()
		// We use separate viper instance to read service-specific flags, and we must "preload" this instance
		// with values we read from config file, otherwise v.Unmarshal will overwrite them with values from flags
		// even if flags were not used.
		var fromConfig map[string]interface{}
		inrec, _ := json.Marshal(targetStruct)
		json.Unmarshal(inrec, &fromConfig)
		for key, val := range fromConfig {
			v.Set(key, val)
		}
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			v.BindPFlag(f.Name, f)
			v.BindEnv(f.Name, "DISPATCH_"+strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)))
		})
		err := v.Unmarshal(targetStruct)
		if err != nil {
			log.Fatalf("Unable to create configuration: %s", err)
		}
	}
}
