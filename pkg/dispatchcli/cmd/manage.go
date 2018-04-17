///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	// The following blank import is to load the gcp plugin (only required to authenticate against GKE clusters)
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mitchellh/go-homedir"
	restclient "k8s.io/client-go/rest"
)

var (
	manageShort = `Manage Dispatch configurations.`
	manageLong  = `Manage Dispatch configurations.`

	manageExample = i18n.T(`
# Enable Dispatch bootstrap mode while specifying bootstrap user
dispatch manage --enable-bootstrap-mode --bootstrap-user admin@example.com -f ./config.yaml

# Enable Dispatch bootstrap mode while specifying service account with public key
dispatch manage --enable-bootstrap-mode --bootstrap-user bootstrap-user --public-key ./bootstrap-user.key.pub -f ./config.yaml

# Disable Dispatch bootstrap mode
dispatch manage --disable-bootstrap-mode -f ./config.yaml`)

	enableBootstrapModeFlag  = false
	disableBootstrapModeFlag = false

	bootstrapUser = ""

	kubeconfigPath = ""
)

const (
	bootstrapSecretName = "dispatch-identity-manager-bootstrap"
)

// NewCmdManage creates a command object for Dispatch "manage" action
func NewCmdManage(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T(`manage [--enable-bootstrap-mode | --disable-bootstrap-mode]`),
		Short:   manageLong,
		Long:    manageLong,
		Example: manageExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := runManage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&installConfigFile, "file", "f", "", "Path to Dispatch install config YAML file")
	cmd.Flags().BoolVarP(&enableBootstrapModeFlag, "enable-bootstrap-mode", "e", false, "enable Dispatch bootstrap mode")
	cmd.Flags().BoolVarP(&disableBootstrapModeFlag, "disable-bootstrap-mode", "d", false, "disable Dispatch bootstrap mode")
	cmd.Flags().StringVar(&bootstrapUser, "bootstrap-user", "", "specify bootstrap user")
	cmd.Flags().StringVar(&publicKeyPath, "public-key", "", "public key file path for bootstrap user (optional)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "customized absolute path to k8s config file (optional)")
	return cmd
}

// prepare k8s client throuhg kubeconfig flag or default kubeconfig location
func prepareK8sClient() (clientset *kubernetes.Clientset, err error) {

	var configPath string
	var config *restclient.Config

	// create k8s config
	if kubeconfigPath != "" {
		configPath, err = homedir.Expand(kubeconfigPath)
	} else {
		homeDir, _ := homedir.Dir()
		configPath = path.Join(homeDir, ".kube", "config")
	}
	config, err = clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load k8s config")
	}

	// create k8s clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create k8s client")
	}
	return clientset, nil
}

func updateBootstrapSecret(secret *v1.Secret, client *kubernetes.Clientset, namesapce string) error {
	var err error
	existed, err := client.CoreV1().Secrets(namesapce).Get(bootstrapSecretName, metav1.GetOptions{})
	if err == nil && existed.Name == bootstrapSecretName {
		_, err = client.CoreV1().Secrets(namesapce).Update(secret)
	} else {
		_, err = client.CoreV1().Secrets(namesapce).Create(secret)
	}
	return err
}

func runManage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	if enableBootstrapModeFlag == disableBootstrapModeFlag {
		runHelp(cmd, []string{})
		return fmt.Errorf("Please specify one of options: --enable-bootstrap-mode or --disable-bootstrap-mode")
	}

	// get namespace from install config
	config, err := readConfig(out, errOut, installConfigFile)
	if err != nil {
		return errors.Wrapf(err, "error reading Dispatch install config")
	}
	namespace := config.DispatchConfig.Chart.Namespace

	// get k8s client
	client, err := prepareK8sClient()
	if err != nil {
		return err
	}

	if enableBootstrapModeFlag {

		// get bootstrap user and publicKey
		if bootstrapUser == "" {
			return fmt.Errorf("Bootstrap user not found, please provide bootstrap user using --bootstrap-user [BOOTSTRAP_USER]")
		}
		// prepare bootstrap secret, data will be base64 encoded by k8s
		data := map[string][]byte{
			"bootstrap_user": []byte(bootstrapUser),
		}

		// get public key if provided
		if publicKeyPath != "" {
			if publicKeyBytes, err := ioutil.ReadFile(publicKeyPath); err == nil {
				encodedPublicKeyString := base64.StdEncoding.EncodeToString(publicKeyBytes)
				data["bootstrap_public_key"] = []byte(encodedPublicKeyString)
			} else {
				return errors.Wrap(err, "Failed to load public key file")
			}
		}

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: bootstrapSecretName,
			},
			Data: data,
		}

		err = updateBootstrapSecret(secret, client, namespace)
		if err != nil {
			return err
		}

		fmt.Println("bootstrap mode enabled, please turn off in production mode")
	}

	if disableBootstrapModeFlag {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: bootstrapSecretName,
			},
		}

		err := updateBootstrapSecret(secret, client, namespace)
		if err != nil {
			return err
		}

		fmt.Println("bootstrap mode disabled")
	}

	return nil
}
