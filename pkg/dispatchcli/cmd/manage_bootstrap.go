package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	// The following blank import is to load GKE auth plugin required when authenticating against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// The following blank import is to load OIDC auth plugin required when authenticating against OIDC-enabledø clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/viper"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/organization"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
)

var (
	bootstrapLong = i18n.T(`Manage bootstrap`)

	disableBootstrapModeFlag = false

	bootstrapSvcAccount = ""
	bootstrapUser       = ""
	bootstrapOrg        = ""
	bootstrapTimeout    time.Duration

	kubeconfigPath = ""

	bootstrapExample = i18n.T(`
# Bootstrap Dispatch - creates a default service account, organization and policies
dispatch manage bootstrap

# Bootstrap Dispatch by specifying a specific bootstrap user and organization name
dispatch manage bootstrap --bootstrap-user admin@example.com --bootstrap-org example-admin-org

# Force disable Dispatch bootstrap mode
dispatch manage bootstrap --disable`)
)

const (
	bootstrapSecretName   = "dispatch-identity-manager-bootstrap"
	defaultSvcAccountName = "default-svc"
	defaultOrgName        = "default"
	defaultPolicyName     = "default-policy"
)

// NewCmdManageBootstrap handles configuration context operations
func NewCmdManageBootstrap(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bootstrap",
		Short:   i18n.T("Bootstrap IAM with an organization, user account and policies"),
		Long:    bootstrapLong,
		Example: bootstrapExample,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"app"},
		Run: func(cmd *cobra.Command, args []string) {
			err := runBootstrap(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&bootstrapUser, "bootstrap-user", "", "specify bootstrap user")
	cmd.Flags().StringVar(&bootstrapSvcAccount, "bootstrap-svc-account", defaultSvcAccountName, "specify bootstrap service account")
	cmd.Flags().StringVar(&bootstrapOrg, "bootstrap-org", defaultOrgName, "specify bootstrap org")
	cmd.Flags().DurationVar(&bootstrapTimeout, "timeout", 2*time.Minute, "specify timeout for checking bootstrap status")
	cmd.Flags().BoolVarP(&disableBootstrapModeFlag, "disable", "d", false, "disable bootstrap mode")
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

func updateBootstrapSecret(secret *kapi.Secret, client *kubernetes.Clientset, namesapce string) error {
	var err error
	existed, err := client.CoreV1().Secrets(namesapce).Get(bootstrapSecretName, metav1.GetOptions{})
	if err == nil && existed.Name == bootstrapSecretName {
		_, err = client.CoreV1().Secrets(namesapce).Update(secret)
	} else {
		_, err = client.CoreV1().Secrets(namesapce).Create(secret)
	}
	return err
}

func disableBootstrapMode(client *kubernetes.Clientset) error {
	secret := &kapi.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: bootstrapSecretName,
		},
	}

	err := updateBootstrapSecret(secret, client, namespace)
	if err != nil {
		return err
	}

	fmt.Println("disabling bootstrap mode (takes up to 30s to take effect)")
	return nil
}

func pemEncodePubKey(key *rsa.PublicKey, path string) ([]byte, error) {

	derEncodedBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal public key in der format")
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derEncodedBytes,
	}
	keyBytes := pem.EncodeToMemory(block)
	if path != "" {
		pubKeyFile, err := os.Create(path)
		os.Chmod(path, 0600)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to write public key to file %s", path))
		}
		pem.Encode(pubKeyFile, block)
	}
	return keyBytes, nil
}

func pemEncodePvtKey(key *rsa.PrivateKey, path string) ([]byte, error) {

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	keyBytes := pem.EncodeToMemory(block)
	if path != "" {
		pvtKeyFile, err := os.Create(path)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to write public key to file %s", path))
		}
		pem.Encode(pvtKeyFile, block)
	}
	return keyBytes, nil
}

func waitForBootstrapStatus(key *rsa.PrivateKey, enable bool) error {
	// TODO: Move to SDK Client for IAM when it's available
	iamClient := identityManagerClient()

	// Set bearer token for bootstrap mode
	if token, err := generateAndSignJWToken("BOOTSTRAP_USER", key, nil); err == nil {
		viperCtx.Set("dispatchToken", token)
	} else {
		return errors.Wrap(err, "failed to generate JWT Token")
	}

	// Wait for bootstrap mode to be enabled
	fmt.Print("waiting for bootstrap status...")
	timer := time.NewTimer(bootstrapTimeout)
	ticker := time.NewTicker(5 * time.Second)
	stopRequest := make(chan bool)
	go func() {
		<-timer.C
		stopRequest <- false
		fmt.Println("timedout")
	}()

	var err error
	go func() {
		for range ticker.C {
			_, err = iamClient.Operations.Home(nil, GetAuthInfoWriter())
			fmt.Print(".")
			if enable && err == nil {
				stopRequest <- true
				fmt.Println("success")
			}
			if !enable && err != nil {
				stopRequest <- true
				fmt.Println("success")
			}
		}
	}()

	success := <-stopRequest
	ticker.Stop()
	timer.Stop()

	if !success {
		return errors.New("failed, please try again or check IAM logs for more information")
	}
	return nil
}

func createSvcAccount() error {
	iamClient := identityManagerClient()

	// Create RSA Key Pair
	svcKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return errors.Wrap(err, "failed to generate RSA key pair")
	}
	svcPubKeyPEM, err := pemEncodePubKey(&svcKey.PublicKey, "")
	svcPubKeyBase64 := base64.StdEncoding.EncodeToString(svcPubKeyPEM)
	serviceAccountModel := &v1.ServiceAccount{
		Name:      &bootstrapSvcAccount,
		PublicKey: &svcPubKeyBase64,
	}
	svcParams := &serviceaccount.AddServiceAccountParams{
		Body:    serviceAccountModel,
		Context: context.Background(),
	}
	fmt.Printf("Creating Service Account: %s\n", bootstrapSvcAccount)
	// Do a force delete if this already exists
	svcDelParams := &serviceaccount.DeleteServiceAccountParams{
		ServiceAccountName: bootstrapSvcAccount,
		Context:            context.Background(),
	}
	_, err = iamClient.Serviceaccount.DeleteServiceAccount(svcDelParams, GetAuthInfoWriter())
	if err != nil {
		if _, ok := err.(*serviceaccount.DeleteServiceAccountNotFound); !ok {
			return errors.Wrap(err, "error deleting existing service account")
		}
	}
	_, err = iamClient.Serviceaccount.AddServiceAccount(svcParams, GetAuthInfoWriter())
	if err != nil {
		return errors.Wrap(err, "error creating service account")
	}

	// Write the private key to config dir
	configFilePath := viper.ConfigFileUsed()
	pvtKeyFilePath := path.Join(filepath.Dir(configFilePath), fmt.Sprintf("%s", bootstrapSvcAccount))
	fmt.Printf("Writing pvt key for service account %s to file %s\n", bootstrapSvcAccount, pvtKeyFilePath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating pvt key file for service account at %s", pvtKeyFilePath))
	}
	_, err = pemEncodePvtKey(svcKey, pvtKeyFilePath)
	if err != nil {
		return err
	}
	fmt.Printf("Setting up CLI current context to use service account %s\n", bootstrapSvcAccount)
	dispatchConfig.ServiceAccount = bootstrapSvcAccount
	dispatchConfig.JWTPrivateKey = pvtKeyFilePath
	return nil
}

func runBootstrap(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	namespace = cmdConfig.Contexts[cmdConfig.Current].Namespace

	// get k8s client
	client, err := prepareK8sClient()
	if err != nil {
		return err
	}

	if disableBootstrapModeFlag {
		return disableBootstrapMode(client)
	}

	// Create RSA Key Pair
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return errors.Wrap(err, "failed to generate RSA key pair")
	}

	publicKeyPEM, err := pemEncodePubKey(&key.PublicKey, "")
	if err != nil {
		return err
	}
	publicKeyBase64Enc := base64.StdEncoding.EncodeToString(publicKeyPEM)
	data := map[string][]byte{
		"bootstrap_user":       []byte("BOOTSTRAP_USER"),
		"bootstrap_public_key": []byte(publicKeyBase64Enc),
	}

	secret := &kapi.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: bootstrapSecretName,
		},
		Data: data,
	}

	err = updateBootstrapSecret(secret, client, namespace)
	if err != nil {
		return err
	}

	fmt.Println("Enabling bootstrap mode")
	waitForBootstrapStatus(key, true)

	// TODO: Move to SDK Client for IAM when it's available
	iamClient := identityManagerClient()

	// Create Organization
	orgModel := &v1.Organization{
		Name: &bootstrapOrg,
	}
	orgParams := &organization.AddOrganizationParams{
		Body:    orgModel,
		Context: context.Background(),
	}
	fmt.Printf("Creating Organization: %s\n", bootstrapOrg)
	_, err = iamClient.Organization.AddOrganization(orgParams, GetAuthInfoWriter())
	if err != nil {
		if _, ok := err.(*organization.AddOrganizationConflict); !ok {
			return errors.Wrap(err, "error creating organization")
		}
	}

	var subject string
	if bootstrapUser == "" {
		err = createSvcAccount()
		if err != nil {
			return err
		}
		subject = bootstrapSvcAccount
	} else {
		subject = bootstrapUser
	}

	// Setup Policies
	policyName := defaultPolicyName
	fmt.Printf("Creating Policy: %s\n", policyName)
	// Deleting any existing policy
	delPolicyParams := &policy.DeletePolicyParams{
		PolicyName: policyName,
		Context:    context.Background(),
	}
	_, err = iamClient.Policy.DeletePolicy(delPolicyParams, GetAuthInfoWriter())
	if err != nil {
		if _, ok := err.(*policy.DeletePolicyNotFound); !ok {
			return errors.Wrap(err, "error deleting existing policy")
		}
	}
	policyModel := &v1.Policy{
		Name: &policyName,
		Rules: []*v1.Rule{
			{
				Subjects:  []string{subject},
				Resources: []string{"*"},
				Actions:   []string{"*"},
			},
		},
	}
	policyParams := &policy.AddPolicyParams{
		Body:    policyModel,
		Context: context.Background(),
	}
	_, err = iamClient.Policy.AddPolicy(policyParams, GetAuthInfoWriter())
	if err != nil {
		return errors.Wrap(err, "error creating policy")
	}

	// write dispatchConfig to file
	cmdConfig.Contexts[cmdConfig.Current] = &dispatchConfig
	vsConfigJSON, err := json.MarshalIndent(cmdConfig, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshalling json")
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), vsConfigJSON, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration to file: %s", viper.ConfigFileUsed())
	}

	// Disable bootstrap mode
	err = disableBootstrapMode(client)
	if err != nil {
		return errors.Wrapf(err, "error disabling bootstrap mode")
	}
	waitForBootstrapStatus(key, false)
	return nil
}
