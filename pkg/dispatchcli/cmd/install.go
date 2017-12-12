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
	"os/exec"
	"path"
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

type installConfig struct {
	Namespace         string          `json:"namespace"`
	Hostname          string          `json:"hostname"`
	Organization      string          `json:"organization"`
	Repository        repostoryConfig `json:"repository"`
	Chart             chartConfig     `json:"chart"`
	CertDir           string          `json:"certificateDirectory"`
	ServiceType       string          `json:"serviceType"`
	HelmRepositoryURL string          `json:"helmRepositoryUrl"`
}

var (
	installLong = `Install the Dispatch framework.`

	installExample    = i18n.T(``)
	installConfigFile = i18n.T(``)
	installServices   = i18n.T(`all`)
	installDryRun     = false
	installDebug      = false

	defaultInstallConfig = installConfig{
		Namespace:         "dispatch",
		Hostname:          "dispatch.vmware.com",
		CertDir:           "/tmp",
		ServiceType:       "NodePort",
		HelmRepositoryURL: "https://s3-us-west-2.amazonaws.com/dispatch-charts",
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

	cmd.Flags().StringVar(&installConfigFile, "file", "", "Path to YAML file")
	cmd.Flags().StringVar(&installServices, "services", "all", "Services to install (defaults to all)")
	cmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Do a dry run, but don't install anything")
	cmd.Flags().BoolVar(&installDebug, "debug", false, "Extra debug output")

	return cmd
}

func makeSSLCert(out, errOut io.Writer, certDir, namespace, domain, certName string) error {
	subject := fmt.Sprintf("/CN=%s/O=%s", domain, domain)
	key := path.Join(certDir, fmt.Sprintf("%s.key", certName))
	cert := path.Join(certDir, fmt.Sprintf("%s.crt", certName))
	openssl := exec.Command(
		"openssl", "req", "-x509", "-nodes", "-days", "365", "-newkey", "rsa:2048",
		"-keyout", key,
		"-out", cert,
		"-subj", subject)
	opensslOut, err := openssl.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, string(opensslOut))
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
	fmt.Println("Copy the following to your $HOME/.dispatch.json")
	fmt.Println(string(b))
	return nil
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
	if installServices == "all" || strings.Contains(installServices, "certs") || !installDryRun {
		err = makeSSLCert(out, errOut, config.CertDir, config.Namespace, config.Hostname, "dispatch-tls")
		if err != nil {
			return errors.Wrapf(err, "Error creating ssl cert %s", installConfigFile)
		}
		err = makeSSLCert(out, errOut, config.CertDir, "kong", "api."+config.Hostname, "api-dispatch-tls")
		if err != nil {
			return errors.Wrapf(err, "Error creating ssl cert %s", installConfigFile)
		}
	}
	err = helmRepoUpdate(out, errOut, "dispatch", config.HelmRepositoryURL)
	if err != nil {
		return errors.Wrapf(err, "Error updating helm")
	}
	if installServices == "all" || strings.Contains(installServices, "ingress") {
		ingressOpts := map[string]string{"controller.service.type": config.ServiceType}
		err = helmInstall(out, errOut, "dispatch/nginx-ingress", "kube-system", "ingress", ingressOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing nginx-ingress chart")
		}
	}
	if installServices == "all" || strings.Contains(installServices, "openfaas") {
		openFaasOpts := map[string]string{"exposeServices": "false"}
		err = helmInstall(out, errOut, "dispatch/openfaas", "openfaas", "openfaas", openFaasOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing openfaas chart")
		}
	}
	if installServices == "all" || strings.Contains(installServices, "api-gateway") {
		kongOpts := map[string]string{"services.proxyService.type": config.ServiceType}
		err = helmInstall(out, errOut, "dispatch/kong", "kong", "api-gateway", kongOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing kong chart")
		}
	}
	if installServices == "all" || strings.Contains(installServices, "dispatch") {
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
		}
		err = helmInstall(out, errOut, "dispatch/dispatch", "dispatch", "dispatch", dispatchOpts)
		if err != nil {
			return errors.Wrapf(err, "Error installing dispatch chart")
		}
	}
	err = writeConfig(out, errOut, config)
	return err
}
