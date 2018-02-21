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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	uninstallLong = `Uninstall the Dispatch framework.`

	uninstallExample         = i18n.T(``)
	uninstallConfigFile      = i18n.T(``)
	uninstallServices        = []string{}
	uninstallDryRun          = false
	uninstallDebug           = false
	uninstallHostname        = i18n.T(``)
	uninstallNamespace       = i18n.T(``)
	uninstallRemoveCertFiles = false
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
			if uninstallConfigFile == "" {
				runHelp(cmd, args)
				return
			}
			err := runUninstall(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&uninstallConfigFile, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringArrayVarP(&uninstallServices, "service", "s", []string{}, "Service to install (defaults to all)")
	cmd.Flags().BoolVar(&uninstallDryRun, "dry-run", false, "Do a dry run, but don't install anything")
	cmd.Flags().BoolVar(&uninstallDebug, "debug", false, "Extra debug output")
	cmd.Flags().BoolVar(&uninstallRemoveCertFiles, "remove-cert-files", false, "Remove the key and certificate files")
	return cmd
}

func uninstallService(service string) bool {
	if len(uninstallServices) == 0 || (len(uninstallServices) == 1 && uninstallServices[0] == "all") {
		return true
	}
	for _, s := range uninstallServices {
		if service == s {
			return true
		}
	}
	return false
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

func helmUninstall(out, errOut io.Writer, namespace, release string, deleteNamespace bool) error {

	args := []string{"delete", "--purge", release}
	if uninstallDebug {
		args = append(args, "--debug")
	}
	if uninstallDryRun {
		args = append(args, "--dry-run")
	}

	fmt.Fprintf(out, "Uninstalling %s from namespace %s\n", release, namespace)
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
	if !uninstallDryRun && deleteNamespace {
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

	config, err := readConfig(out, errOut, uninstallConfigFile)
	if err != nil {
		return err
	}

	if uninstallDebug {
		b, _ := json.MarshalIndent(config, "", "    ")
		fmt.Fprintln(out, string(b))
	}

	configDir, err := homedir.Expand(configDest)

	if uninstallService("certs") || !uninstallDryRun {
		err = uninstallSSLCert(out, errOut, configDir, config.DispatchConfig.Chart.Namespace, config.DispatchConfig.Host, config.DispatchConfig.TLS.SecretName)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling ssl cert %s", uninstallConfigFile)
		}
		err = uninstallSSLCert(out, errOut, configDir, config.APIGateway.Chart.Namespace, config.APIGateway.Host, config.APIGateway.TLS.SecretName)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling ssl cert %s", uninstallConfigFile)
		}
	}
	if uninstallService("ingress") {
		err = helmUninstall(out, errOut, config.Ingress.Chart.Namespace, config.Ingress.Chart.Release, false)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling nginx-ingress chart")
		}
	}
	if uninstallService("postgres") {
		err = helmUninstall(out, errOut, config.PostgresConfig.Chart.Namespace, config.PostgresConfig.Chart.Release, false)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling postgres chart")
		}
	}
	if uninstallService("openfaas") {
		err = helmUninstall(out, errOut, config.OpenFaas.Chart.Namespace, config.OpenFaas.Chart.Release, true)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling openfaas chart")
		}
	}
	if uninstallService("riff") {
		err = helmUninstall(out, errOut, config.Riff.Chart.Namespace, config.Riff.Chart.Release, true)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling riff chart")
		}
	}
	if uninstallService("kafka") {
		err = helmUninstall(out, errOut, config.Kafka.Chart.Namespace, config.Kafka.Chart.Release, true)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling kafka chart")
		}
	}
	if uninstallService("api-gateway") {
		err = helmUninstall(out, errOut, config.APIGateway.Chart.Namespace, config.APIGateway.Chart.Release, true)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling kong chart")
		}
	}
	if uninstallService("dispatch") {
		err = helmUninstall(out, errOut, config.DispatchConfig.Chart.Namespace, config.DispatchConfig.Chart.Release, true)
		if err != nil {
			return errors.Wrapf(err, "Error uninstalling dispatch chart")
		}
	}
	return err
}
