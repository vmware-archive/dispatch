///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	uninstallLong = `Uninstall the Dispatch framework.`

	uninstallExample         = i18n.T(``)
	uninstallConfigFile      = i18n.T(``)
	uninstallServices        []string
	uninstallDryRun          = false
	uninstallDebug           = false
	uninstallKeepNS          = false
	uninstallRemoveCertFiles = false
	uninstallSingleNS        = ""
	uninstallHelmTillerNS    = ""

	serviceNamespaces map[string]string
	namespaceCount    map[string]int
)

// NewCmdUninstall creates a command object for the uninstallation of dispatch
// compontents
func NewCmdUninstall(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uninstall [flags]",
		Short:   i18n.T("Uninstall some or all of dispatch"),
		Long:    uninstallLong,
		Example: uninstallExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := runUninstall(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&uninstallConfigFile, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringArrayVarP(&uninstallServices, "service", "s", []string{}, "Service to uninstall (defaults to all)")
	cmd.Flags().BoolVar(&uninstallDryRun, "dry-run", false, "Do a dry run, but don't install anything")
	cmd.Flags().BoolVar(&uninstallDebug, "debug", false, "Extra debug output")
	cmd.Flags().BoolVar(&uninstallRemoveCertFiles, "remove-cert-files", false, "Remove the key and certificate files")
	cmd.Flags().StringVar(&uninstallSingleNS, "single-namespace", "", "If specified, all dispatch components will be uninstalled from that namespace")
	cmd.Flags().StringVar(&uninstallHelmTillerNS, "tiller-namespace", "kube-system", "The namespace where Helm's tiller has been installed")
	cmd.Flags().BoolVar(&uninstallKeepNS, "keep-namespaces", false, "Keep namespaces (do not delete them together with services)")
	return cmd
}

func uninstallService(service string) (bool, bool) {
	ns, hasNS := serviceNamespaces[service]
	lastInNamespace := false
	if ns != "kube-system" && hasNS {
		count := namespaceCount[ns]
		if count <= 1 {
			lastInNamespace = true
		}
	}

	if len(uninstallServices) == 0 || (len(uninstallServices) == 1 && uninstallServices[0] == "all") {
		if hasNS {
			namespaceCount[ns]--
		}
		return lastInNamespace, true
	}
	for _, s := range uninstallServices {
		if service == s {
			if hasNS {
				namespaceCount[ns]--
			}
			return lastInNamespace, true
		}
	}
	return false, false
}

func uninstallSSLCert(out, errOut io.Writer, configDir, namespace, domain, certName string) error {
	key := path.Join(configDir, fmt.Sprintf("%s.key", domain))
	cert := path.Join(configDir, fmt.Sprintf("%s.crt", domain))
	var err error
	if uninstallRemoveCertFiles {
		if err = os.Remove(key); err != nil {
			return errors.Wrapf(err, "Failed to remove file %s", key)
		}
	}
	if uninstallRemoveCertFiles {
		if err = os.Remove(cert); err != nil {
			return errors.Wrapf(err, "Failed to remove file %s", key)
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
	return nil
}

func helmUninstall(out, errOut io.Writer, service, release string, deleteNamespace bool) error {
	namespace := serviceNamespaces[service]

	args := []string{"delete", "--tiller-namespace", uninstallHelmTillerNS, "--purge", release}
	if uninstallDebug {
		args = append(args, "--debug")
	}
	if uninstallDryRun {
		args = append(args, "--dry-run")
	}

	if namespace != "" {
		fmt.Fprintf(out, "Uninstalling %s from namespace %s\n", release, namespace)
	} else {
		fmt.Fprintf(out, "Uninstalling %s\n", release)
	}
	helm := exec.Command("helm", args...)
	helmOut, err := helm.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(helmOut), "not found") {
			return errors.Wrapf(err, string(helmOut))
		}
	}
	if uninstallDebug {
		fmt.Fprintln(out, string(helmOut))
	}
	if uninstallKeepNS {
		// if explicitly asked, keep namespace in every case
		deleteNamespace = false
	}
	if !uninstallDryRun && deleteNamespace {
		fmt.Fprintf(out, "Removing namespace %s\n", namespace)
		kubectl := exec.Command(
			"kubectl", "delete", "namespace", namespace)
		kubectlOut, err := kubectl.CombinedOutput()
		if err != nil {
			if !strings.Contains(string(kubectlOut), "NotFound") {
				return errors.Wrapf(err, string(kubectlOut))
			}
		}
	}
	return nil
}

func runUninstall(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	var err error
	config := &installConfig{}
	err = yaml.Unmarshal([]byte(defaultInstallConfigYaml), config)
	if err != nil {
		return errors.Wrapf(err, "error decoding default install config yaml file")
	}

	serviceNamespaces = map[string]string{
		"dispatch":        config.DispatchConfig.Chart.Namespace,
		"api-gateway":     config.APIGateway.Chart.Namespace,
		"postgres":        config.PostgresConfig.Chart.Namespace,
		"openfaas":        config.OpenFaas.Chart.Namespace,
		"kubeless":        config.Kubeless.Chart.Namespace,
		"riff":            config.Riff.Chart.Namespace,
		"ingress":         config.Ingress.Chart.Namespace,
		"docker-registry": config.DockerRegistry.Chart.Namespace,
		"service-catalog": config.DispatchConfig.Service.K8sServiceCatalog.Namespace,
		"rabbitmq":        config.RabbitMQ.Chart.Namespace,
		"kafka":           config.Kafka.Chart.Namespace,
		"jaeger":          config.Jaeger.Chart.Namespace,
	}

	if uninstallSingleNS != "" {
		for k := range serviceNamespaces {
			serviceNamespaces[k] = uninstallSingleNS
		}
	} else {
		config, err = readConfig(out, errOut, uninstallConfigFile)
		if err != nil {
			return err
		}
	}

	namespaceCount = make(map[string]int)
	for _, n := range serviceNamespaces {
		if _, ok := namespaceCount[n]; !ok {
			namespaceCount[n] = 1
			continue
		}
		namespaceCount[n]++
	}

	if uninstallDebug {
		b, _ := json.MarshalIndent(config, "", "    ")
		fmt.Fprintln(out, string(b))
	}

	configDir, err := homedir.Expand(configDest)

	if _, ok := uninstallService("certs"); ok || !uninstallDryRun {
		err = uninstallSSLCert(out, errOut, configDir, config.DispatchConfig.Chart.Namespace, config.DispatchConfig.Host, config.DispatchConfig.TLS.SecretName)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling ssl cert %s", uninstallConfigFile)
		}
		err = uninstallSSLCert(out, errOut, configDir, config.APIGateway.Chart.Namespace, config.APIGateway.Host, config.APIGateway.TLS.SecretName)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling ssl cert %s", uninstallConfigFile)
		}
		err = helmUninstall(out, errOut, "lets-encrypt", config.LetsEncrypt.Chart.Release, false)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling certificate chart")
		}
	}
	if removeNS, ok := uninstallService("ingress"); ok {
		err = helmUninstall(out, errOut, "ingress", config.Ingress.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling nginx-ingress chart")
		}
	}
	if removeNS, ok := uninstallService("postgres"); ok {
		err = helmUninstall(out, errOut, "postgres", config.PostgresConfig.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling postgres chart")
		}
	}
	if removeNS, ok := uninstallService("docker-registry"); ok {
		err = helmUninstall(out, errOut, "docker-registry", config.DockerRegistry.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling openfaas chart")
		}
	}
	if removeNS, ok := uninstallService("openfaas"); ok {
		err = helmUninstall(out, errOut, "openfaas", config.OpenFaas.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling openfaas chart")
		}
	}
	if removeNS, ok := uninstallService("jaeger"); ok {
		err = helmUninstall(out, errOut, "jaeger", config.Jaeger.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling jaeger chart")
		}
	}
	if removeNS, ok := uninstallService("riff"); ok {
		err = helmUninstall(out, errOut, "riff", config.Riff.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling riff chart")
		}
	}
	if removeNS, ok := uninstallService("kafka"); ok {
		err = helmUninstall(out, errOut, "kafka", config.Kafka.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling kafka chart")
		}
	}
	if removeNS, ok := uninstallService("rabbitmq"); ok {
		err = helmUninstall(out, errOut, "rabbitmq", config.RabbitMQ.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling rabbitmq chart")
		}
	}
	if removeNS, ok := uninstallService("api-gateway"); ok {
		err = helmUninstall(out, errOut, "api-gateway", config.APIGateway.Chart.Release, removeNS)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling kong chart")
		}
	}
	if _, ok := uninstallService("dispatch"); ok {
		err = helmUninstall(out, errOut, "dispatch", config.DispatchConfig.Chart.Release, true)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling dispatch chart")
		}
	}

	return err
}
