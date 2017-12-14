///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

type repostoryConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type imageConfig struct {
	Host string `json:"host"`
	Tag  string `json:"tag"`
}

type chartConfig struct {
	Image imageConfig `json:"image"`
}

type apiGatewayConfig struct {
	ServiceType string `json:"serviceType"`
}

type oauth2ProxyConfig struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	CookieSecret string `json:"cookieSecret"`
}

type installConfig struct {
	Namespace         string            `json:"namespace"`
	Hostname          string            `json:"hostname"`
	Organization      string            `json:"organization"`
	Repository        repostoryConfig   `json:"repository"`
	Chart             chartConfig       `json:"chart"`
	CertDir           string            `json:"certificateDirectory"`
	ServiceType       string            `json:"serviceType"`
	APIGateway        apiGatewayConfig  `json:"apiGateway"`
	HelmRepositoryURL string            `json:"helmRepositoryUrl"`
	PersistData       bool              `json:"persistData"`
	ConfigDest        string            `json:"configDest"`
	OAuth2Proxy       oauth2ProxyConfig `json:"oauth2Proxy"`
}

var (
	installLong = `Install the Dispatch framework.`

	installExample    = i18n.T(``)
	installConfigFile = i18n.T(`dispatch`)
	installServices   = []string{}
	chartsDir         = i18n.T(``)
	installDryRun     = false
	installDebug      = false
	configDest        = i18n.T(``)

	defaultInstallConfig = installConfig{
		Namespace:         "dispatch",
		Hostname:          "dispatch.vmware.com",
		CertDir:           "/tmp",
		ServiceType:       "NodePort",
		HelmRepositoryURL: "https://s3-us-west-2.amazonaws.com/dispatch-charts",
		PersistData:       false,
		APIGateway: apiGatewayConfig{
			ServiceType: "NodePort",
		},
		OAuth2Proxy: oauth2ProxyConfig{
			ClientID:     "invalid",
			ClientSecret: "invalid",
			// note: this secret is just a placeholder, don't use this one!
			CookieSecret: "YVBLBQXd4CZo1vnUTSM/3w==",
		},
	}
)

// NewCmdInstall creates a command object for the generic "get" action, which
// retrieves one or more resources from a server.
func NewCmdInstall(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install [flags]",
		Short:   i18n.T("Display one or many resources"),
		Long:    installLong,
		Example: installExample,
		Run: func(cmd *cobra.Command, args []string) {
			if installConfigFile == "" {
				runHelp(cmd, args)
				return
			}
			err := runInstall(out, errOut, cmd, args)
			CheckErr(err)
		},
		SuggestFor: []string{"list"},
	}

	cmd.Flags().StringVarP(&installConfigFile, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringArrayVarP(&installServices, "service", "s", []string{}, "Service to install (defaults to all)")
	cmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Do a dry run, but don't install anything")
	cmd.Flags().BoolVar(&installDebug, "debug", false, "Extra debug output")
	cmd.Flags().StringVar(&chartsDir, "charts-dir", "dispatch", "File path to local charts (for chart development)")
	cmd.Flags().StringVarP(&configDest, "destination", "d", "", "Destination of the CLI configuration")
	return cmd
}

func makeSSLCert(out, errOut io.Writer, certDir, namespace, domain, certName string) error {
	subject := fmt.Sprintf("/CN=%s/O=%s", domain, domain)
	key := path.Join(certDir, fmt.Sprintf("%s.key", certName))
	cert := path.Join(certDir, fmt.Sprintf("%s.crt", certName))
	var err error
	// If cert and key exist, reuse them
	if _, err = os.Stat(key); os.IsNotExist(err) {
		if _, err = os.Stat(cert); os.IsNotExist(err) {
			openssl := exec.Command(
				"openssl", "req", "-x509", "-nodes", "-days", "365", "-newkey", "rsa:2048",
				"-keyout", key,
				"-out", cert,
				"-subj", subject)
			opensslOut, err := openssl.CombinedOutput()
			if err != nil {
				return errors.Wrapf(err, string(opensslOut))
			}
		}
	}
	kubectl := exec.Command(
		"kubectl", "delete", "secret", "tls", certName, "-n", namespace)
	kubectlOut, err := kubectl.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(kubectlOut), "NotFound") {
			return errors.Wrapf(err, string(kubectlOut))
		}
	}
	kubectl = exec.Command(
		"kubectl", "create", "namespace", namespace)
	kubectlOut, err = kubectl.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(kubectlOut), "AlreadyExists") {
			return errors.Wrapf(err, string(kubectlOut))
		}
	}
	kubectl = exec.Command(
		"kubectl", "create", "secret", "tls", certName, "-n", namespace, "--key", key, "--cert", cert)
	kubectlOut, err = kubectl.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(kubectlOut))
	}
	return nil
}

func helmRepoUpdate(out, errOut io.Writer, name, repoURL string) error {
	helm := exec.Command(
		"helm", "repo", "add", name, repoURL)
	helmOut, err := helm.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	helm = exec.Command("helm", "repo", "update")
	helmOut, err = helm.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	return nil
}

func helmDepUp(out, errOut io.Writer, chart string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error getting current working directory")
	}
	err = os.Chdir(chart)
	if err != nil {
		return errors.Wrap(err, "Error changing directory")
	}
	helm := exec.Command("helm", "dep", "up")
	helmOut, err := helm.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	return os.Chdir(cwd)
}

func helmInstall(out, errOut io.Writer, chart, namespace, release string, options map[string]string) error {

	args := []string{"upgrade", release, chart, "--install", "--namespace", namespace}
	for k, v := range options {
		args = append(args, "--set", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, "--debug")
	args = append(args, "--wait")
	if installDryRun {
		args = append(args, "--dry-run")
	}

	if installDebug {
		fmt.Printf("debug: helm")
		for _, a := range args {
			fmt.Printf(" %s", a)
		}
		fmt.Printf("\n")
	}

	helm := exec.Command("helm", args...)
	helmOut, err := helm.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	if installDebug {
		fmt.Println(string(helmOut))
	}
	return nil
}

func writeConfig(out, errOut io.Writer, config installConfig) error {
	dispatchConfig.Organization = config.Organization
	dispatchConfig.Host = config.Hostname
	dispatchConfig.Port = 443
	b, err := json.MarshalIndent(dispatchConfig, "", "    ")
	if err != nil {
		return err
	}
	if configDest == "" {
		fmt.Println("Copy the following to your $HOME/.dispatch.json")
		fmt.Println(string(b))
	} else {
		configPath := path.Join(configDest, ".dispatch.json")
		fmt.Printf("Config file written to: %s", configPath)
		return ioutil.WriteFile(configPath, b, 0644)
	}
	return nil
}

func installService(service string) bool {
	if len(installServices) == 0 {
		return true
	}
	for _, s := range installServices {
		if service == s {
			return true
		}
	}
	return false
}

func runInstall(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	// Default Config
	config := defaultInstallConfig
	b, err := ioutil.ReadFile(installConfigFile)
	if err != nil {
		return errors.Wrapf(err, "Error reading file %s", installConfigFile)
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return errors.Wrapf(err, "Error decoding yaml file %s", installConfigFile)
	}
	if installService("certs") || !installDryRun {
		err = makeSSLCert(out, errOut, config.CertDir, config.Namespace, config.Hostname, "dispatch-tls")
		if err != nil {
			return errors.Wrapf(err, "Error creating ssl cert %s", installConfigFile)
		}
		err = makeSSLCert(out, errOut, config.CertDir, "kong", "api."+config.Hostname, "api-dispatch-tls")
		if err != nil {
			return errors.Wrapf(err, "Error creating ssl cert %s", installConfigFile)
		}
	}
	if chartsDir == "dispatch" {
		err = helmRepoUpdate(out, errOut, chartsDir, config.HelmRepositoryURL)
		if err != nil {
			return errors.Wrapf(err, "Error updating helm")
		}
	}
	if installService("ingress") {
		chart := path.Join(chartsDir, "nginx-ingress")
		ingressOpts := map[string]string{
			"controller.service.type": config.ServiceType,
		}
		err = helmInstall(out, errOut, chart, "kube-system", "ingress", ingressOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing nginx-ingress chart")
		}
	}
	if installService("openfaas") {
		chart := path.Join(chartsDir, "openfaas")
		openFaasOpts := map[string]string{"exposeServices": "false"}
		err = helmInstall(out, errOut, chart, "openfaas", "openfaas", openFaasOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing openfaas chart")
		}
	}
	if installService("api-gateway") {
		chart := path.Join(chartsDir, "kong")
		kongOpts := map[string]string{
			"services.proxyService.type": config.APIGateway.ServiceType,
		}
		err = helmInstall(out, errOut, chart, "kong", "api-gateway", kongOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing kong chart")
		}
	}
	if installService("dispatch") {
		chart := path.Join(chartsDir, "dispatch")
		if chartsDir != "dispatch" {
			err = helmDepUp(out, errOut, chart)
			if err != nil {
				return errors.Wrap(err, "Error updating chart dependencies")
			}
		}
		openFaasAuth := fmt.Sprintf(
			`{"username":"%s","password":"%s","email":"%s"}`,
			config.Repository.Username,
			config.Repository.Password,
			config.Repository.Email)
		openFaasAuthEncoded := base64.StdEncoding.EncodeToString([]byte(openFaasAuth))
		dispatchOpts := map[string]string{
			"function-manager.faas.openfaas.registryAuth":  openFaasAuthEncoded,
			"function-manager.faas.openfaas.imageRegistry": config.Repository.Host,
			"global.host":                                  config.Hostname,
			"global.image.host":                            config.Chart.Image.Host,
			"global.image.tag":                             config.Chart.Image.Tag,
			"global.debug":                                 "true",
			"global.data.persist":                          strconv.FormatBool(config.PersistData),
			"oauth2-proxy.app.clientID":                    config.OAuth2Proxy.ClientID,
			"oauth2-proxy.app.clientSecret":                config.OAuth2Proxy.ClientSecret,
			"oauth2-proxy.app.cookieSecret":                config.OAuth2Proxy.CookieSecret,
		}
		err = helmInstall(out, errOut, chart, "dispatch", "dispatch", dispatchOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing dispatch chart")
		}
	}
	err = writeConfig(out, errOut, config)
	return err
}
