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
	"github.com/vmware/dispatch/pkg/utils"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var cmdConfig struct {
	Current  string                 `json:"current"`
	Contexts map[string]*hostConfig `json:"contexts"`
}

type hostConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Scheme         string `json:"scheme"`
	Organization   string `json:"organization"`
	Cookie         string `json:"cookie"`
	Insecure       bool   `json:"insecure"`
	Namespace      string `json:"namespace,omitempty"`
	JSON           bool   `json:"-"`
	APIHTTPSPort   int    `json:"api-https-port,omitempty"`
	APIHTTPPort    int    `json:"api-http-port,omitempty"`
	Token          string `json:"-"`
	ServiceAccount string `json:"serviceaccount,omitempty"`
	JWTPrivateKey  string `json:"jwtprivatekey,omitempty"`
}

// Current Config Context
var dispatchConfig = hostConfig{}

var validResources = i18n.T(`Valid resource types include:
	* apis
	* applications
	* base-images
	* eventdrivers
	* eventdrivertypes
	* functions
	* images
	* secrets
	* subscriptions
    `)

var (
	dispatchConfigPath = ""

	cmdFlagApplication = i18n.T(``)

	cmds *cobra.Command

	// Holds the config map of the current context
	viperCtx *viper.Viper
)

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
	// Initialize the config map
	cmdConfig.Contexts = make(map[string]*hostConfig)
	viper.Unmarshal(&cmdConfig)

	viperCtx = viper.Sub(fmt.Sprintf("contexts.%s", cmdConfig.Current))
	if viperCtx != nil {
		viperCtx.BindPFlag("host", cmds.PersistentFlags().Lookup("host"))
		viperCtx.BindPFlag("port", cmds.PersistentFlags().Lookup("port"))
		viperCtx.BindPFlag("organization", cmds.PersistentFlags().Lookup("organization"))
		viperCtx.BindPFlag("insecure", cmds.PersistentFlags().Lookup("insecure"))
		viperCtx.BindPFlag("json", cmds.PersistentFlags().Lookup("json"))
		viperCtx.BindPFlag("token", cmds.PersistentFlags().Lookup("token"))
		viperCtx.BindPFlag("serviceAccount", cmds.PersistentFlags().Lookup("service-account"))
		viperCtx.BindPFlag("jwtPrivateKey", cmds.PersistentFlags().Lookup("jwt-private-key"))
		// Limited support for env variables
		viperCtx.BindEnv("config", "DISPATCH_CONFIG")
		viperCtx.BindEnv("insecure", "DISPATCH_INSECURE")
		viperCtx.BindEnv("token", "DISPATCH_TOKEN")
		viperCtx.BindEnv("organization", "DISPATCH_ORGANIZATION")
		viperCtx.BindEnv("serviceAccount", "DISPATCH_SERVICE_ACCOUNT")
		viperCtx.BindEnv("jwtPrivateKey", "DISPATCH_JWT_PRIVATE_KEY")
		viperCtx.Unmarshal(&dispatchConfig)
	}
}

// NewCLI creates cobra object for top-level Dispatch CLI
func NewCLI(in io.Reader, out, errOut io.Writer) *cobra.Command {
	// Parent command to which all subcommands are added.
	cobra.OnInitialize(initConfig)
	cmds = &cobra.Command{
		Use:   "dispatch",
		Short: i18n.T("dispatch allows to interact with VMware Dispatch framework."),
		Long:  i18n.T("dispatch allows to interact with VMware Dispatch framework."),
		Run:   runHelp,
	}
	cmds.PersistentFlags().StringVar(&dispatchConfigPath, "config", "", "config file (default is $HOME/.dispatch)")
	cmds.PersistentFlags().String("host", "dispatch.example.com", "Dispatch host to connect to")
	cmds.PersistentFlags().Int("port", 443, "Port which Dispatch is listening on")
	cmds.PersistentFlags().String("organization", "", "Organization name")
	cmds.PersistentFlags().Bool("insecure", false, "If true, will ignore verifying the server's certificate and your https connection is insecure.")
	cmds.PersistentFlags().BoolVar(&dispatchConfig.JSON, "json", false, "Output raw JSON")
	cmds.PersistentFlags().String("token", "", "JWT Bearer Token")
	cmds.PersistentFlags().String("service-account", "", "Name of the service account, if specified, a jwt-private-key is also required")
	cmds.PersistentFlags().String("jwt-private-key", "", "JWT private key file path")

	cmds.AddCommand(NewCmdGet(out, errOut))
	cmds.AddCommand(NewCmdCreate(out, errOut))
	cmds.AddCommand(NewCmdUpdate(out, errOut))
	cmds.AddCommand(NewCmdExec(out, errOut))
	cmds.AddCommand(NewCmdDelete(out, errOut))
	cmds.AddCommand(NewCmdLogin(in, out, errOut))
	cmds.AddCommand(NewCmdLogout(in, out, errOut))
	cmds.AddCommand(NewCmdEmit(out, errOut))
	cmds.AddCommand(NewCmdInstall(out, errOut))
	cmds.AddCommand(NewCmdUninstall(out, errOut))
	cmds.AddCommand(NewCmdVersion(out))
	cmds.AddCommand(NewCmdIam(out, errOut))
	cmds.AddCommand(NewCmdManage(out, errOut))
	cmds.AddCommand(NewCmdLog(out, errOut))
	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func resourceName(name string) string {
	if name != "" {
		return name
	}
	return utils.RandomResourceName()

}

func getOrgFromConfig() string {
	if dispatchConfig.Organization == "" {
		fatal(fmt.Sprintf("error: missing organization. Please specify it using --organization flag or set in the config file %s. "+
			"If this is a new Dispatch installation, you can check `dispatch manage bootstrap --help` command for initial setup.", viper.ConfigFileUsed()), 1)
	}
	return dispatchConfig.Organization
}
