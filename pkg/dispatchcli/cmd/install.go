///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"crypto/rand"
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
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/imdario/mergo"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

type repostoryConfig struct {
	Host     string `json:"host,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

type imageConfig struct {
	Host string `json:"host,omitempty"`
	Tag  string `json:"tag,omitempty"`
}

type chartConfig struct {
	Image *imageConfig `json:"image,omitempty"`
}

type apiGatewayConfig struct {
	ServiceType string `json:"serviceType,omitempty"`
}

type oauth2ProxyConfig struct {
	ClientID     string `json:"clientID,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	CookieSecret string `json:"cookieSecret,omitmepty"`
}

type installConfig struct {
	Namespace         string             `json:"namespace,omitempty"`
	Hostname          string             `json:"hostname,omitempty"`
	Organization      string             `json:"organization,omitempty"`
	Repository        *repostoryConfig   `json:"repository,omitempty"`
	Chart             *chartConfig       `json:"chart,omitempty"`
	ServiceType       string             `json:"serviceType,omitempty"`
	APIGateway        *apiGatewayConfig  `json:"apiGateway,omitempty"`
	HelmRepositoryURL string             `json:"helmRepositoryUrl,omitempty"`
	PersistData       bool               `json:"persistData"`
	OAuth2Proxy       *oauth2ProxyConfig `json:"oauth2Proxy"`
}

var installSchema = `
{
	"type": "object",
	"properties": {
		"namespace": {
			"type": "string",
			"pattern": "^[a-z0-9][a-z0-9\\-\\.]*[a-z0-9]$",
			"maxLength": 253,
			"default": "dispatch"
		},
		"hostname": {
			"type": "string",
			"pattern": "^[a-z0-9][a-z0-9\\-\\.]*[a-z0-9]$",
			"maxLength": 63,
			"default": "dispatch.vmware.com"
		},
		"organization": {
			"type": "string"
		},
		"serviceType": {
			"type": "string",
			"enum": ["NodePort", "LoadBalancer", "ClusterIP"],
			"default": "NodePort"
		},
		"helmRepostoryUrl": {
			"type": "string",
			"format": "uri",
			"default": "https://s3-us-west-2.amazonaws.com/dispatch-charts"
		},
		"persistData": {
			"type": "boolean",
			"default": false
		},
		"repository": {
			"type": "object",
			"properties": {
				"host": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"email": {
					"type": "string",
					"format": "email"
				},
				"username": {
					"type": "string"
				}
			},
			"required": ["host", "password", "email", "username"]
		},
		"chart": {
			"type": "object",
			"properties": {
				"image": {
					"type": "object",
					"properties": {
						"host": {
							"type": "string"
						},
						"tag": {
							"type": "string"
						}
					}
				}
			}
		},
		"apiGateway": {
			"type": "object",
			"properties": {
				"serviceType": {
					"type": "string",
					"enum": ["NodePort", "LoadBalancer", "ClusterIP"],
					"default": "NodePort"
				}
			}
		},
		"oauth2Proxy": {
			"type": "object",
			"properties": {
				"clientID": {
					"type": "string"
				},
				"clientSecret": {
					"type": "string"
				},
				"cookieSecret": {
					"type": "string"
				}
			},
			"required": ["clientID", "clientSecret"]
		}
	},
	"required": ["oauth2Proxy", "repository"]
}
`

var (
	installLong = `Install the Dispatch framework.`

	installExample    = i18n.T(``)
	installConfigFile = i18n.T(``)
	installServices   = []string{}
	chartsDir         = i18n.T(``)
	installDryRun     = false
	installDebug      = false
	configDest        = i18n.T(``)
	servicePort       = 443

	defaultInstallConfig = installConfig{
		Namespace:         "dispatch",
		Organization:      "dispatch",
		Hostname:          "dispatch.vmware.com",
		ServiceType:       "NodePort",
		HelmRepositoryURL: "https://s3-us-west-2.amazonaws.com/dispatch-charts",
		PersistData:       false,
		APIGateway: &apiGatewayConfig{
			ServiceType: "NodePort",
		},
	}
)

// NewCmdInstall creates a command object for the generic "get" action, which
// retrieves one or more resources from a server.
func NewCmdInstall(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install [flags]",
		Short:   i18n.T("Install some or all of dispatch"),
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
	}

	cmd.Flags().StringVarP(&installConfigFile, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringArrayVarP(&installServices, "service", "s", []string{}, "Service to install (defaults to all)")
	cmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Do a dry run, but don't install anything")
	cmd.Flags().BoolVar(&installDebug, "debug", false, "Extra debug output")
	cmd.Flags().StringVar(&chartsDir, "charts-dir", "dispatch", "File path to local charts (for chart development)")
	cmd.Flags().StringVarP(&configDest, "destination", "d", "~/.dispatch", "Destination of the CLI configuration")
	return cmd
}

func makeSSLCert(out, errOut io.Writer, configDir, namespace, domain, certName string) error {
	subject := fmt.Sprintf("/CN=%s/O=%s", domain, domain)
	key := path.Join(configDir, fmt.Sprintf("%s.key", domain))
	cert := path.Join(configDir, fmt.Sprintf("%s.crt", domain))
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
		fmt.Fprintf(out, "debug: helm")
		for _, a := range args {
			fmt.Fprintf(out, " %s", a)
		}
		fmt.Fprintf(out, "\n")
	}

	helm := exec.Command("helm", args...)
	helmOut, err := helm.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	if installDebug {
		fmt.Fprintln(out, string(helmOut))
	}
	return nil
}

func writeConfig(out, errOut io.Writer, configDir string, config installConfig) error {
	dispatchConfig.Organization = config.Organization
	dispatchConfig.Host = config.Hostname
	dispatchConfig.Port = servicePort
	b, err := json.MarshalIndent(dispatchConfig, "", "    ")
	if err != nil {
		return err
	}
	if installDryRun {
		fmt.Fprintf(out, "Copy the following to your %s/config.json\n", configDir)
		fmt.Fprintln(out, string(b))
	} else {
		configPath := path.Join(configDir, "config.json")
		fmt.Fprintf(out, "Config file written to: %s\n", configPath)
		return ioutil.WriteFile(configPath, b, 0644)
	}
	return nil
}

func installService(service string) bool {
	if len(installServices) == 0 || (len(installServices) == 1 && installServices[0] == "all") {
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
	b, err := ioutil.ReadFile(installConfigFile)
	if err != nil {
		return errors.Wrapf(err, "Error reading file %s", installConfigFile)
	}
	config := installConfig{}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return errors.Wrapf(err, "Error decoding yaml file %s", installConfigFile)
	}

	v := new(spec.Schema)
	err = json.Unmarshal([]byte(installSchema), v)
	if err != nil {
		panic(err)
	}
	err = validate.AgainstSchema(v, config, strfmt.Default)
	if err != nil {
		return errors.Wrapf(err, "Configuration error")
	}

	err = mergo.Merge(&config, defaultInstallConfig)
	if err != nil {
		return errors.Wrapf(err, "Error merging default values")
	}

	if installDebug {
		b, _ = json.MarshalIndent(config, "", "    ")
		fmt.Fprintln(out, string(b))
	}

	configDir, err := homedir.Expand(configDest)
	if !installDryRun {
		_, err = os.Stat(configDir)
		if os.IsNotExist(err) {
			err = os.MkdirAll(configDir, 0755)
			if err != nil {
				return errors.Wrapf(err, "Error creating config destination directory")
			}
		}
	}

	if installService("certs") || !installDryRun {
		err = makeSSLCert(out, errOut, configDir, config.Namespace, config.Hostname, "dispatch-tls")
		if err != nil {
			return errors.Wrapf(err, "Error creating ssl cert %s", installConfigFile)
		}
		err = makeSSLCert(out, errOut, configDir, "kong", "api."+config.Hostname, "api-dispatch-tls")
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
		if config.ServiceType == "NodePort" {
			kubectl := exec.Command(
				"kubectl", "get", "svc", "ingress-nginx-ingress-controller", "-n", "kube-system", "-o", "jsonpath={.spec.ports[?(@.name==\"https\")].nodePort}")
			kubectlOut, err := kubectl.CombinedOutput()
			if err != nil {
				return errors.Wrapf(err, string(kubectlOut))
			}
			servicePort, err = strconv.Atoi(string(kubectlOut))
			if err != nil {
				return errors.Wrapf(err, "Error fetching nginx-ingress node port")
			}
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
		// Resets the cookie every deployment if not specified
		if config.OAuth2Proxy.CookieSecret == "" {
			cookie := make([]byte, 16)
			_, err := rand.Read(cookie)
			if err != nil {
				return errors.Wrap(err, "Error creating cookie secret")
			}
			config.OAuth2Proxy.CookieSecret = base64.StdEncoding.EncodeToString(cookie)
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
			"global.port":                                  strconv.Itoa(servicePort),
			"global.debug":                                 "true",
			"global.data.persist":                          strconv.FormatBool(config.PersistData),
			"oauth2-proxy.app.clientID":                    config.OAuth2Proxy.ClientID,
			"oauth2-proxy.app.clientSecret":                config.OAuth2Proxy.ClientSecret,
			"oauth2-proxy.app.cookieSecret":                config.OAuth2Proxy.CookieSecret,
		}
		// If unset values default to chart values
		if config.Chart != nil && config.Chart.Image != nil {
			if config.Chart.Image.Host != "" {
				dispatchOpts["global.image.host"] = config.Chart.Image.Host
			}
			if config.Chart.Image.Tag != "" {
				dispatchOpts["global.image.tag"] = config.Chart.Image.Tag
			}
		}
		if installDebug {
			for k, v := range dispatchOpts {
				fmt.Fprintf(out, "%v: %v\n", k, v)
			}
		}
		err = helmInstall(out, errOut, chart, "dispatch", "dispatch", dispatchOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing dispatch chart")
		}
	}
	err = writeConfig(out, errOut, configDir, config)
	return err
}
