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
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/utils"
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
	Output         string `json:"-"`
	APIHTTPSPort   int    `mapstructure:"api-https-port" json:"api-https-port,omitempty"`
	APIHTTPPort    int    `mapstructure:"api-http-port" json:"api-http-port,omitempty"`
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

	execRunner Runner
)

func initConfig() {
	// Don't forget to read config either from dispatchConfigPath or from home directory!
	if dispatchConfigPath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(dispatchConfigPath)
		readConfigFile(false)
	} else if config := os.Getenv("DISPATCH_CONFIG"); config != "" {
		viper.SetConfigFile(config)
		readConfigFile(false)
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
		readConfigFile(true)
	}
	// TODO (bjung): add config command to print config used
	// Initialize the config map
	cmdConfig.Contexts = make(map[string]*hostConfig)
	viper.Unmarshal(&cmdConfig)

	if viperCtx := viper.Sub(fmt.Sprintf("contexts.%s", cmdConfig.Current)); viperCtx != nil {
		bindFlagsAndEnv(viperCtx)
		viperCtx.Unmarshal(&dispatchConfig)
	} else {
		// If there is no context, just read value from the flags
		v := viper.New()
		bindFlagsAndEnv(v)
		v.Unmarshal(&dispatchConfig)
	}
	if dispatchConfig.Output != "" {
		matched, _ := regexp.MatchString("yaml|json", dispatchConfig.Output)
		if !matched {
			fmt.Println("invalid format option [yaml|json]")
			os.Exit(1)
		}
	}
	if dispatchConfig.APIHTTPPort == 0 {
		dispatchConfig.APIHTTPPort = utils.DefaultAPIHTTPPort
	}
	if dispatchConfig.APIHTTPSPort == 0 {
		dispatchConfig.APIHTTPSPort = utils.DefaultAPIHTTPSPort
	}
}

func readConfigFile(ignoreNotFound bool) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ignoreNotFound && ok {
			return
		}
		fatal(fmt.Sprintf("Error reading config file %s: %s", viper.ConfigFileUsed(), err), 1)
	}
}

func writeConfigFile() {
	writeConfigFile := viper.ConfigFileUsed()
	if writeConfigFile == "" {
		// User probably did not use any config files - use default location
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		writeConfigFile = path.Join(home, ".dispatch", "config.json")
	}
	if cmdConfig.Current == "" {
		cmdConfig.Current = formatContextName(dispatchConfig.Host)
	}
	cmdConfig.Contexts[cmdConfig.Current] = &dispatchConfig
	vsConfigJSON, err := json.MarshalIndent(cmdConfig, "", "    ")
	if err != nil {
		fatal(fmt.Sprintf("error writing configuration, %s", "error marshalling json"), 1)
	}

	err = ioutil.WriteFile(writeConfigFile, vsConfigJSON, 0644)
	if err != nil {
		fatal(fmt.Sprintf("error writing configuration to file: %s, %s", viper.ConfigFileUsed(), err), 1)
	}
}

func bindFlagsAndEnv(v *viper.Viper) {
	cmds.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		key := utils.SeparatedToCamelCase(f.Name, "-")
		v.BindPFlag(key, f)
		v.BindEnv(key, "DISPATCH_"+strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)))
	})
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

	execRunner = &execCmdRunner{}

	cmds.PersistentFlags().StringVar(&dispatchConfigPath, "config", "", "config file (default is $HOME/.dispatch)")
	cmds.PersistentFlags().String("host", "localhost", "Dispatch host to connect to")
	cmds.PersistentFlags().Int("port", 443, "Port which Dispatch is listening on")
	cmds.PersistentFlags().String("scheme", "https", "The protocol scheme to use, either http or https")
	cmds.PersistentFlags().String("organization", "", "Organization name")
	cmds.PersistentFlags().Bool("insecure", false, "If true, will ignore verifying the server's certificate and your https connection is insecure.")
	cmds.PersistentFlags().BoolVar(&dispatchConfig.JSON, "json", false, "Output raw JSON")
	cmds.PersistentFlags().StringVarP(&dispatchConfig.Output, "output", "o", "", "Output format [json|yaml]")
	cmds.PersistentFlags().String("token", "", "JWT Bearer Token")
	cmds.PersistentFlags().String("service-account", "", "Name of the service account, if specified, a jwt-private-key is also required")
	cmds.PersistentFlags().String("jwt-private-key", "", "JWT private key file path")

	cmds.PersistentFlags().MarkHidden("json")

	cmds.AddCommand(NewCmdApply(out, errOut))
	cmds.AddCommand(NewCmdGet(out, errOut))
	cmds.AddCommand(NewCmdCreate(out, errOut))
	cmds.AddCommand(NewCmdUpdate(out, errOut))
	cmds.AddCommand(NewCmdEdit(out, errOut))
	cmds.AddCommand(NewCmdExec(out, errOut))
	cmds.AddCommand(NewCmdDelete(out, errOut))
	cmds.AddCommand(NewCmdLogin(in, out, errOut))
	cmds.AddCommand(NewCmdLogout(in, out, errOut))
	cmds.AddCommand(NewCmdEmit(out, errOut))
	cmds.AddCommand(NewCmdVersion(out))
	cmds.AddCommand(NewCmdIam(out, errOut))
	cmds.AddCommand(NewCmdContext(out, errOut))
	cmds.AddCommand(NewCmdCompletion(out))
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
	return dispatchConfig.Organization
}

func formatOutput(out io.Writer, list bool, models interface{}) (bool, error) {
	v := reflect.ValueOf(models)
	if !list && (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) {
		item := v.Index(0)
		models = item.Interface()
	}
	if dispatchConfig.JSON || dispatchConfig.Output == "json" {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return true, encoder.Encode(models)
	}
	if dispatchConfig.Output == "yaml" {
		encoder := yaml.NewEncoder(out)
		return true, encoder.Encode(models)
	}
	return false, nil
}
