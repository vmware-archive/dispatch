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
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/validator.v9"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	dispatchHelmRepositoryURL = "https://s3-us-west-2.amazonaws.com/dispatch-charts"
	dispatchHost              = ""
	dispatchHostIP            = ""
	apiGatewayHost            = ""
	apiGatewayHostIP          = ""
	servicesAvailable         = []string{"certs", "ingress", "postgres", "api-gateway", "kafka", "rabbitmq", "docker-registry", "dispatch"}
	servicesEnabled           = map[string]bool{}
	certReqSANIP              = "subjectAltName = IP:%s"
	certReqSANDNS             = "subjectAltName = DNS:%s"
	certReqTemplate           = `
[req]
req_extensions = v3_req
distinguished_name = dn

[dn]

[v3_req]
basicConstraints = CA:TRUE
keyUsage = digitalSignature, keyEncipherment
`
)

type chartConfig struct {
	Chart     string            `json:"chart,omitempty" validate:"required"`
	Namespace string            `json:"namespace,omitempty" validate:"required,hostname"`
	Release   string            `json:"release,omitempty" validate:"required"`
	Repo      string            `json:"repo,omitempty" validate:"omitempty,uri"`
	Version   string            `json:"version,omitempty" validate:"omitempty"`
	Opts      map[string]string `json:"opts,omitempty" validate:"omitempty"`
}

type dockerRegistry struct {
	Chart *chartConfig `json:"chart,omitempty" validate:"required"`
}

type ingressConfig struct {
	Chart       *chartConfig `json:"chart,omitempty" validate:"required"`
	ServiceType string       `json:"serviceType,omitempty" validate:"required,eq=NodePort|eq=LoadBalancer|eq=ClusterIP"`
}

type postgresConfig struct {
	Chart       *chartConfig `json:"chart,omitempty" validate:"required"`
	Database    string       `json:"database,omitempty" validate:"required"`
	Username    string       `json:"username,omitempty" validate:"required"`
	Password    string       `json:"password,omitempty" validate:"required"`
	Host        string       `json:"host,omitempty" validate:"required,hostname"`
	Port        int          `json:"port,omitempty" validate:"required"`
	Persistence bool         `json:"persistence,omitempty" validate:"omitempty"`
}

type rabbitMQConfig struct {
	Chart       *chartConfig `json:"chart,omitempty" validate:"required"`
	Username    string       `json:"username,omitempty" validate:"required"`
	Password    string       `json:"password,omitempty" validate:"required"`
	Persistence bool         `json:"persistence,omitempty"`
	Host        string       `json:"host,omitempty" validate:"omitempty,hostname"`
	Port        int          `json:"port,omitempty" validate:"required"`
}

type tlsConfig struct {
	CertFile   string `json:"certFile,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	SecretName string `json:"secretName,omitempty" validate:"required"`
	Ca         string `json:"ca,omitempty" validate:"eq=letsEncrypt|eq="`
	Insecure   bool   `json:"insecure,omitempty" validate:"omitempty"`
}

type apiGatewayConfig struct {
	Chart       *chartConfig `json:"chart,omitempty" validate:"required"`
	ServiceType string       `json:"serviceType,omitempty" validate:"required,eq=NodePort|eq=LoadBalancer|eq=ClusterIP"`
	Database    string       `json:"database,omitempty" validate:"required"`
	Host        string       `json:"host,omitempty" validate:"required,hostname|ip"`
	TLS         *tlsConfig   `json:"tls,omitempty" validate:"required"`
}

type openfaasConfig struct {
	Chart         *chartConfig `json:"chart,omitempty" validate:"required"`
	ExposeService bool         `json:"exposeService,omitempty" validate:"omitempty"`
}

type kubelessConfig struct {
	Chart *chartConfig `json:"chart,omitempty" validate:"required"`
}

type riffConfig struct {
	Chart *chartConfig `json:"chart,omitempty" validate:"required"`
}

type kafkaConfig struct {
	Chart          *chartConfig `json:"chart,omitempty" validate:"required"`
	Brokers        []string     `json:"brokers,omitempty" validate:"required"`
	ZookeeperNodes []string     `json:"zookeeperNodes,omitempty" validate:"required"`
}

type jaegerConfig struct {
	Chart   *chartConfig `json:"chart,omitempty" validate:"required"`
	Agent   string       `json:"agent,omitempty" validate:"omitempty"`
	Enabled bool         `json:"enabled,omitempty" validate:"omitempty"`
}

type certManagerConfig struct {
	Chart   *chartConfig `json:"chart,omitempty" validate:"required"`
	Enabled bool         `json:"enabled,omitempty" validate:"omitempty"`
}

type clouddnsConfig struct {
	Project    string `json:"project" validate:"required"`
	SecretName string `json:"secretName" validate:"required"`
	SecretKey  string `json:"secretKey" validate:"required"`
}

type route53Config struct {
	AccessKeyID string `json:"accessKeyID" validate:"required"`
	SecretName  string `json:"secretName" validate:"required"`
	SecretKey   string `json:"secretKey" validate:"required"`
}

type dnsConfig struct {
	Provider string          `json:"provider,omitempty" validate:"omitempty"`
	Route53  *route53Config  `json:"route53,omitempty" validate:"required"`
	Clouddns *clouddnsConfig `json:"clouddns,omitempty" validate:"required"`
}

type letsEncryptConfig struct {
	Chart   *chartConfig `json:"chart,omitempty" validate:"required"`
	Email   string       `json:"email,omitempty" validate:"required,email"`
	Staging bool         `json:"staging,omitempty" validate:"omitempty"`
	DNS     *dnsConfig   `json:"dns,omitempty" validate:"required"`
}

type imageConfig struct {
	Host string `json:"host,omitempty" validate:"omitempty"`
	Tag  string `json:"tag,omitempty"  validate:"omitempty"`
}
type oauth2ProxyConfig struct {
	Provider      string `json:"provider,omitempty" validate:"omitempty"`
	OIDCIssuerURL string `json:"oidcIssuerURL,omitempty" validate:"omitempty"`
	ClientID      string `json:"clientID,omitempty" validate:"required"`
	ClientSecret  string `json:"clientSecret,omitempty" validate:"required"`
	CookieSecret  string `json:"cookieSecret,omitempty" validate:"omitempty"`
}
type imageRegistryConfig struct {
	Name     string `json:"name,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"omitempty"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Username string `json:"username,omitempty" validate:"omitempty"`
	Insecure bool   `json:"insecure,omitempty" validate:"omitempty"`
}
type k8sServiceCatalogConfig struct {
	Namespace string `json:"namespace,omitempty" validate:"required"`
}
type serviceCatalogConfig struct {
	Catalog           string                   `json:"catalog,omitemtpy" validate:"required,eq=k8sservicecatalog"`
	K8sServiceCatalog *k8sServiceCatalogConfig `json:"k8sservicecatalog,omitempty"`
}
type dispatchInstallConfig struct {
	Chart           *chartConfig          `json:"chart,omitempty" validate:"required"`
	Host            string                `json:"host,omitempty" validate:"required,hostname|ip"`
	Port            int                   `json:"port,omitempty" validate:"required"`
	Image           *imageConfig          `json:"image,omitempty" validate:"omitempty"`
	Debug           bool                  `json:"debug,omitempty" validate:"omitempty"`
	Trace           bool                  `json:"trace,omitempty" validate:"omitempty"`
	Database        string                `json:"database,omitempty" validate:"required,eq=postgres"`
	PersistData     bool                  `json:"persistData,omitempty" validate:"omitempty"`
	ImageRegistry   *imageRegistryConfig  `json:"imageRegistry,omitempty" validate:"omitempty"`
	ImagePullSecret string                `json:"imagePullSecret,omitempty" validate:"omitempty"`
	OAuth2Proxy     *oauth2ProxyConfig    `json:"oauth2Proxy,omitempty" validate:"required"`
	TLS             *tlsConfig            `json:"tls,omitempty" validate:"required"`
	SkipAuth        bool                  `json:"skipAuth,omitempty" validate:"omitempty"`
	Faas            string                `json:"faas,omitempty" validate:"required,eq=openfaas|eq=riff|eq=kubeless"`
	EventTransport  string                `json:"eventTransport,omitempty" validate:"required,eq=kafka|eq=rabbitmq"`
	Service         *serviceCatalogConfig `json:"service,omitemtpy" validate:"required"`
}

type installConfig struct {
	HelmRepositoryURL string                 `json:"helmRepositoryUrl,omitempty" validate:"required,uri"`
	Ingress           *ingressConfig         `json:"ingress,omitempty" validate:"required"`
	PostgresConfig    *postgresConfig        `json:"postgresql,omitempty" validate:"required"`
	APIGateway        *apiGatewayConfig      `json:"apiGateway,omitempty" validate:"required"`
	OpenFaas          *openfaasConfig        `json:"openfaas,omitempty" validate:"required"`
	Kubeless          *kubelessConfig        `json:"kubeless,omitempty" validate:"required"`
	Riff              *riffConfig            `json:"riff,omitempty" validate:"required"`
	Kafka             *kafkaConfig           `json:"kafka,omitempty" validate:"required"`
	RabbitMQ          *rabbitMQConfig        `json:"rabbitmq,omitempty" validate:"required"`
	DispatchConfig    *dispatchInstallConfig `json:"dispatch,omitempty" validate:"required"`
	DockerRegistry    *dockerRegistry        `json:"dockerRegistry,omitempty" validate:"omitempty"`
	Jaeger            *jaegerConfig          `json:"jaeger,omitempty" validate:"required"`
	CertManager       *certManagerConfig     `json:"certManager,omitempty" validate:"required"`
	LetsEncrypt       *letsEncryptConfig     `json:"letsEncrypt,omitempty" validate:"omitempty"`
}

var (
	installLong = `Install the Dispatch framework.`

	installExample    = i18n.T(``)
	installConfigFile = i18n.T(``)
	installServices   []string
	installChartsDir  = i18n.T(``)
	installChartsRepo = i18n.T(``)
	installSingleNS   = ""
	installDryRun     = false
	installDebug      = false
	configDest        = i18n.T(``)
	helmTimeout       = 600
	helmIgnoreCheck   = false
	helmTillerNS      = ""
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
	cmd.Flags().StringArrayVarP(&installServices, "service", "s", []string{}, "Service to install (defaults to all). Add 'no-' prefix to service name to disable it.")
	cmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Do a dry run, but don't install anything")
	cmd.Flags().BoolVar(&installDebug, "debug", false, "Extra debug output")
	cmd.Flags().StringVar(&installChartsRepo, "charts-repo", "dispatch", "Helm Chart Repo used")
	cmd.Flags().StringVar(&installChartsDir, "charts-dir", "dispatch", "File path to local charts (for chart development)")
	cmd.Flags().StringVar(&installSingleNS, "single-namespace", "", "If specified, all dispatch components will be installed to that namespace")
	cmd.Flags().StringVarP(&configDest, "destination", "d", "~/.dispatch", "Destination of the CLI configuration")
	cmd.Flags().IntVarP(&helmTimeout, "timeout", "t", 600, "Timeout (in seconds) passed to Helm when installing charts")
	cmd.Flags().StringVar(&helmTillerNS, "tiller-namespace", "kube-system", "The namespace where Helm's tiller has been installed")
	cmd.Flags().BoolVar(&helmIgnoreCheck, "ignore-helm-check", false, "Ignore checking for failed Helm releases")
	return cmd
}

func createCertConfFile(domain, path string) error {
	var san string
	if ip := net.ParseIP(domain); ip != nil {
		san = fmt.Sprintf(certReqSANIP, domain)
	} else {
		san = fmt.Sprintf(certReqSANDNS, domain)
	}
	certContent := []byte(fmt.Sprintf("%s\n%s", certReqTemplate, san))
	if err := ioutil.WriteFile(path, certContent, 0644); err != nil {
		return errors.Wrapf(err, "error saving cert configuration file")
	}
	return nil
}

func installCert(out, errOut io.Writer, configDir, namespace, domain string, tls *tlsConfig) (bool, error) {
	var key, cert string
	var insecure bool

	if tls.CertFile != "" {
		if tls.PrivateKey == "" {
			return false, errors.New("error installing certificate: missing private key for the tls cert")
		}
		key = tls.PrivateKey
		cert = tls.CertFile
	} else {
		// make a new key and cert.
		fmt.Fprintf(out, "Creating new certificate for domain %s\n", domain)
		domainShort := domain
		if len(domain) > 64 {
			fmt.Fprintf(errOut, "WARNING: Domain %s is longer than 64 characters, the certificate common name will be truncated", domain)
			domainShort = domain[0:63]
		}

		certReqFile := path.Join(configDir, fmt.Sprintf("%s.cnf", domainShort))
		subject := fmt.Sprintf("/CN=%s/O=%s", domainShort, domainShort)
		key = path.Join(configDir, fmt.Sprintf("%s.key", domainShort))
		cert = path.Join(configDir, fmt.Sprintf("%s.crt", domainShort))
		var err error
		// If cert and key exist, reuse them
		if _, err = os.Stat(key); os.IsNotExist(err) {
			if _, err = os.Stat(cert); os.IsNotExist(err) {
				createCertConfFile(domain, certReqFile)
				args := []string{"req", "-x509", "-nodes", "-days", "365", "-newkey", "rsa:2048",
					"-config", certReqFile,
					"-keyout", key,
					"-out", cert,
					"-subj", subject}
				if installDebug {
					fmt.Fprintf(out, "debug: openssl")
					for _, a := range args {
						fmt.Fprintf(out, " %s", a)
					}
					fmt.Fprintf(out, "\n")
				}
				opensslOut, err := execRunner.Run("openssl", args...)
				if err != nil {
					return false, errors.Wrapf(err, string(opensslOut))
				}
			}
		}
		// The cert is self-signed and therefore will not validate, so set the insecure flag
		insecure = true
	}
	fmt.Fprintf(out, "Updating certificate in namespace %s\n", namespace)
	kubectlOut, err := execRunner.Run(
		"kubectl", "delete", "secret", "tls", tls.SecretName, "-n", namespace)
	if err != nil {
		if !strings.Contains(string(kubectlOut), "NotFound") {
			return insecure, errors.Wrapf(err, string(kubectlOut))
		}
	}

	err = createNamespace(namespace)
	if err != nil {
		return insecure, err
	}

	kubectlOut, err = execRunner.Run(
		"kubectl", "create", "secret", "tls", tls.SecretName, "-n", namespace, "--key", key, "--cert", cert)
	if err != nil {
		return insecure, errors.Wrapf(err, string(kubectlOut))
	}
	return insecure, nil
}

func helmRepoUpdate(out, errOut io.Writer, name, repoURL string) error {
	helmOut, err := execRunner.Run(
		"helm", "repo", "add", name, repoURL)
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	helmOut, err = execRunner.Run("helm", "repo", "update")
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	return nil
}

func helmDepUp(out, errOut io.Writer, chart string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "error getting current working directory")
	}
	err = os.Chdir(chart)
	if err != nil {
		return errors.Wrap(err, "error changing directory")
	}
	helmOut, err := execRunner.Run("helm", "dep", "up")
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	return os.Chdir(cwd)
}

func helmCheckFailedRelease(out, errOut io.Writer) error {
	args := []string{"list", "--failed", "--short", "--tiller-namespace", helmTillerNS}
	if installSingleNS != "" {
		args = append(args, "--namespace", installSingleNS)
	}
	helmOut, err := execRunner.Run("helm", args...)
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	if string(helmOut) != "" {
		fmt.Fprintf(errOut, "Error: Following failed helm releases were found:\n%s", string(helmOut))
		return errors.New("please delete the failed releases using 'helm delete --purge <release_name>' and re-run the dispatch install command")
	}
	return nil
}

func helmInstall(out, errOut io.Writer, meta *chartConfig) error {

	repo := ""
	chart := meta.Chart
	if meta.Repo != "" {
		// if user specified a repo, use that
		repo = meta.Repo
	} else if installChartsDir == "dispatch" {
		// use dispatch chart repo
		repo = dispatchHelmRepositoryURL
	} else {
		// use the local charts
		chart = path.Join(installChartsDir, meta.Chart)
	}

	namespace := meta.Namespace

	args := []string{"upgrade", "-i", meta.Release, chart, "--namespace", namespace, "--tiller-namespace", helmTillerNS}

	for k, v := range meta.Opts {
		args = append(args, "--set", fmt.Sprintf("%s=%s", k, v))
	}

	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if meta.Version != "" {
		args = append(args, "--version", meta.Version)
	}
	if installDebug {
		args = append(args, "--debug")
	}
	args = append(args, "--wait")
	args = append(args, "--timeout", strconv.Itoa(helmTimeout))
	if installDryRun {
		args = append(args, "--dry-run")
	}
	if installDebug {
		fmt.Fprintf(out, "debug: helm")
		for _, a := range args {
			fmt.Fprintf(out, " %s", a)
		}
		fmt.Fprintf(out, "\n")
	} else {
		fmt.Fprintf(out, "Installing %s helm chart\n", chart)
	}

	helmOut, err := execRunner.Run("helm", args...)
	if err != nil {
		return errors.Wrapf(err, string(helmOut))
	}
	if installDebug {
		fmt.Fprintln(out, string(helmOut))
	}
	fmt.Fprintf(out, "Successfully installed %s chart - %s\n", chart, meta.Release)
	return nil
}

func writeConfig(out, errOut io.Writer, configDir string, config *installConfig) error {

	c := &hostConfig{}
	c.Host = dispatchHost
	if dispatchHost == "" {
		c.Host = dispatchHostIP
	}

	c.Port = config.DispatchConfig.Port
	c.Insecure = config.DispatchConfig.TLS.Insecure
	c.Namespace = config.DispatchConfig.Chart.Namespace
	c.APIHTTPPort = dispatchConfig.APIHTTPPort
	c.APIHTTPSPort = dispatchConfig.APIHTTPSPort

	if config.DispatchConfig.SkipAuth {
		// In SkipAuth mode, a dummy org is required to be passed as HEADER to avoid swagger runtime returning an
		// HTTP 422
		c.Organization = "dispatch"
	}

	if config.APIGateway.ServiceType == "NodePort" {
		fmt.Fprintf(out, "dispatch api-gateway is running at http port: %d and https port: %d\n",
			c.APIHTTPPort, c.APIHTTPSPort)
	}
	cmdConfig.Current = formatContextName(c.Host)
	cmdConfig.Contexts[cmdConfig.Current] = c
	b, err := json.MarshalIndent(cmdConfig, "", "    ")
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

func readConfig(out, errOut io.Writer, file string) (*installConfig, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading file %s", file)
	}
	config := installConfig{}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding yaml file %s", file)
	}

	defaultInstallConfig := installConfig{}
	err = yaml.Unmarshal([]byte(defaultInstallConfigYaml), &defaultInstallConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding default install config yaml file")
	}
	if installDebug {
		b, _ := json.MarshalIndent(config, "", "    ")
		fmt.Fprintln(out, "Config before merge")
		fmt.Fprintln(out, string(b))
	}
	err = mergo.Merge(&config, defaultInstallConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "Error merging default values")
	}
	return &config, nil
}

// selectServices configures services to install
func selectServices(out io.Writer, config *installConfig) error {
	for _, service := range servicesAvailable {
		servicesEnabled[service] = true
	}

	servicesEnabled[config.DispatchConfig.Faas] = true

	// Reset rabbitmq && kafka services, and enable them again based on install config.
	servicesEnabled["kafka"] = false
	if config.DispatchConfig.EventTransport == "kafka" || config.DispatchConfig.Faas == "riff" {
		servicesEnabled["kafka"] = true
	}

	servicesEnabled["rabbitmq"] = false
	if config.DispatchConfig.EventTransport == "rabbitmq" {
		servicesEnabled["rabbitmq"] = true
	}

	// TODO: once Jaeger becomes a default service, remove this
	servicesEnabled["jaeger"] = config.Jaeger.Enabled

	// Most used combination - all default services enabled
	if len(installServices) == 0 || installServices[0] == "all" {
		return nil
	}

	// We have two modes: whitelisting or blacklisting. Adding "no-" prefix to the service name
	// enters blacklist mode. Modes cannot be mixed.
	var whitelistMode, blacklistMode bool

	for _, service := range installServices {
		if strings.HasPrefix(service, "no-") {
			if whitelistMode {
				return errors.New("can either whitelist or blacklist services, not both")
			}

			blacklistMode = true
			service := strings.TrimPrefix(service, "no-")
			if _, ok := servicesEnabled[service]; !ok {
				return fmt.Errorf("unknown service: %s", service)
			}
			servicesEnabled[service] = false
		} else {
			if blacklistMode {
				return errors.New("can either whitelist or blacklist services, not both")
			}

			// whitelist mode
			if !whitelistMode {
				// we entered whitelist mode, during first encounter we need to reset enabled services
				for service := range servicesEnabled {
					servicesEnabled[service] = false
				}
				whitelistMode = true
			}
			if _, ok := servicesEnabled[service]; !ok {
				return fmt.Errorf("unknown service: %s", service)
			}
			servicesEnabled[service] = true
		}
	}
	if installDebug {
		fmt.Fprint(out, "\nServices to be installed:\n")
		for service, enabled := range servicesEnabled {
			if enabled {
				fmt.Fprintf(out, "* %s\n", service)
			}
		}
	}
	return nil
}

func installService(service string) bool {
	return servicesEnabled[service]
}

func getK8sServiceNodePort(service, namespace string, https bool) (int, error) {

	standardPort := 80
	if https {
		standardPort = 443
	}

	kubectlOut, err := execRunner.Run(
		"kubectl", "get", "svc", service, "-n", namespace,
		"-o", fmt.Sprintf("jsonpath={.spec.ports[?(@.port==%d)].nodePort}", standardPort))

	if err != nil {
		return -1, errors.Wrapf(err, string(kubectlOut))
	}
	nodePort, err := strconv.Atoi(string(kubectlOut))
	if err != nil {
		return -1, errors.Wrapf(err, "Error fetching node port")
	}
	return nodePort, nil
}

func getK8sServiceClusterIP(service, namespace string) (string, error) {

	kubectlOut, err := execRunner.Run(
		"kubectl", "get", "svc", service, "-n", namespace,
		"-o", fmt.Sprintf("jsonpath={.spec.clusterIP}"))

	if err != nil {
		return "", errors.Wrapf(err, string(kubectlOut))
	}
	return string(kubectlOut), nil
}

func getK8sServiceLoadBalancerIPorHost(service, namespace string) (string, error) {

	kubectlOut, err := execRunner.Run(
		"kubectl", "get", "svc", service, "-n", namespace,
		"-o", fmt.Sprintf("jsonpath={.status.loadBalancer.ingress[].ip}"))

	if err != nil {
		return "", errors.Wrapf(err, string(kubectlOut))
	}
	ip := string(kubectlOut)
	if ip == "" {
		kubectlOut, err := execRunner.Run(
			"kubectl", "get", "svc", service, "-n", namespace,
			"-o", fmt.Sprintf("jsonpath={.status.loadBalancer.ingress[].hostname}"))

		if err != nil {
			return "", errors.Wrapf(err, string(kubectlOut))
		}
		return string(kubectlOut), nil
	}
	return ip, nil
}

func installCertManager(out, errOut io.Writer, config *installConfig) error {
	err := helmInstall(out, errOut, config.CertManager.Chart)
	if err != nil {
		return errors.Wrapf(err, "Error installing cert-manager chart")
	}
	return nil
}

func createNamespace(ns string) error {
	// We can't rely on "kubectl create" returning "AlreadyExists" to check for the namespace existence,
	// as platforms with limited privileges may return "Forbidden" even namespace already exists.
	_, err := execRunner.Run("kubectl", "get", "namespace", ns)
	if err != nil {
		//
		kubectlOut, err := execRunner.Run("kubectl", "create", "namespace", ns)
		if err != nil {
			return errors.Wrapf(err, "Error creating namespace: %s: %s", ns, string(kubectlOut))
		}
	}
	return nil
}

func createCertManagerCert(out, errOut io.Writer, config *installConfig) error {
	letsEncryptOpts := map[string]string{
		"staging":               strconv.FormatBool(config.LetsEncrypt.Staging),
		"dispatch.secretName":   config.DispatchConfig.TLS.SecretName,
		"dispatch.dnsName":      dispatchHost,
		"dispatch.namespace":    config.DispatchConfig.Chart.Namespace,
		"apiGateway.secretName": config.APIGateway.TLS.SecretName,
		"apiGateway.dnsName":    apiGatewayHost,
		"apiGateway.namespace":  config.APIGateway.Chart.Namespace,
		"acmeEmail":             config.LetsEncrypt.Email,
	}
	switch p := config.LetsEncrypt.DNS.Provider; p {
	case "route53":
		letsEncryptOpts["provider"] = config.LetsEncrypt.DNS.Provider
		letsEncryptOpts["route53.accessKeyID"] = config.LetsEncrypt.DNS.Route53.AccessKeyID
		letsEncryptOpts["route53.secretName"] = config.LetsEncrypt.DNS.Route53.SecretName
		letsEncryptOpts["route53.secretKey"] = config.LetsEncrypt.DNS.Route53.SecretKey
	case "clouddns":
		letsEncryptOpts["provider"] = config.LetsEncrypt.DNS.Provider
		letsEncryptOpts["clouddns.project"] = config.LetsEncrypt.DNS.Clouddns.Project
		letsEncryptOpts["clouddns.secretName"] = config.LetsEncrypt.DNS.Clouddns.SecretName
		letsEncryptOpts["clouddns.secretKey"] = config.LetsEncrypt.DNS.Clouddns.SecretKey
	}
	mergo.Merge(&config.LetsEncrypt.Chart.Opts, letsEncryptOpts)

	err := createNamespace(config.DispatchConfig.Chart.Namespace)
	if err != nil {
		return err
	}
	err = createNamespace(config.APIGateway.Chart.Namespace)
	if err != nil {
		return err
	}

	err = helmInstall(out, errOut, config.LetsEncrypt.Chart)
	if err != nil {
		return errors.Wrapf(err, "Error installing certificate chart")
	}
	return nil
}

func runInstall(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	config, err := readConfig(out, errOut, installConfigFile)
	if err != nil {
		return err
	}

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return errors.Wrapf(err, "Configuration error")
	}

	if config.HelmRepositoryURL != "" {
		dispatchHelmRepositoryURL = config.HelmRepositoryURL
	}

	if !helmIgnoreCheck {
		// Until https://github.com/kubernetes/helm/issues/3353 is resolved we need to make sure
		// there are no failed helm releases.
		err := helmCheckFailedRelease(out, errOut)
		if err != nil {
			return errors.Wrapf(err, "Helm check failed")
		}
	}

	var ip net.IP
	if ip = net.ParseIP(config.DispatchConfig.Host); ip != nil {
		// User specified an IP address for dispatch host.
		dispatchHostIP = ip.String()
	} else {
		dispatchHost = config.DispatchConfig.Host
	}
	if ip = net.ParseIP(config.APIGateway.Host); ip != nil {
		// User specified an IP address for dispatch host.
		apiGatewayHostIP = ip.String()
	} else {
		apiGatewayHost = config.APIGateway.Host
	}

	if installDebug {
		b, _ := json.MarshalIndent(config, "", "    ")
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

	selectServices(out, config)

	// Override the default namespace for all charts
	if installSingleNS != "" {
		config.DispatchConfig.Chart.Namespace = installSingleNS
		config.APIGateway.Chart.Namespace = installSingleNS
		config.PostgresConfig.Chart.Namespace = installSingleNS
		config.OpenFaas.Chart.Namespace = installSingleNS
		config.Kubeless.Chart.Namespace = installSingleNS
		config.Riff.Chart.Namespace = installSingleNS
		config.Ingress.Chart.Namespace = installSingleNS
		config.DockerRegistry.Chart.Namespace = installSingleNS
		config.DispatchConfig.Service.K8sServiceCatalog.Namespace = installSingleNS
		config.RabbitMQ.Chart.Namespace = installSingleNS
		config.Kafka.Chart.Namespace = installSingleNS
		config.Kafka.Brokers = []string{fmt.Sprintf("transport-kafka.%s:9092", installSingleNS)}
	}

	if installChartsDir == "dispatch" {
		err = helmRepoUpdate(out, errOut, installChartsDir, config.HelmRepositoryURL)
		if err != nil {
			return errors.Wrapf(err, "Error updating helm")
		}
	}

	if installService("certs") && !installDryRun {
		if config.DispatchConfig.TLS.Ca == "letsEncrypt" {
			err = installCertManager(out, errOut, config)
			if err != nil {
				return err
			}
			err = createCertManagerCert(out, errOut, config)
			if err != nil {
				return err
			}
		} else {
			insecure, err := installCert(out, errOut, configDir, config.DispatchConfig.Chart.Namespace, config.DispatchConfig.Host, config.DispatchConfig.TLS)
			if err != nil {
				return errors.Wrapf(err, "Error installing cert for %s", config.DispatchConfig.Host)
			}
			if insecure {
				config.DispatchConfig.TLS.Insecure = insecure
			}
			_, err = installCert(out, errOut, configDir, config.APIGateway.Chart.Namespace, config.APIGateway.Host, config.APIGateway.TLS)
			if err != nil {
				return errors.Wrapf(err, "Error installing  cert for %s", config.APIGateway.Host)
			}
		}
	}

	if installService("ingress") {
		ingressOpts := map[string]string{
			"controller.service.type": config.Ingress.ServiceType,
		}
		mergo.Merge(&config.Ingress.Chart.Opts, ingressOpts)
		err = helmInstall(out, errOut, config.Ingress.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing nginx-ingress chart")
		}
		service := fmt.Sprintf("%s-nginx-ingress-controller", config.Ingress.Chart.Release)
		if config.Ingress.ServiceType == "NodePort" {
			config.DispatchConfig.Port, err = getK8sServiceNodePort(service, config.Ingress.Chart.Namespace, true)
			if err != nil {
				return err
			}
		} else if config.Ingress.ServiceType == "LoadBalancer" {
			config.DispatchConfig.Host, err = getK8sServiceLoadBalancerIPorHost(service, config.Ingress.Chart.Namespace)
			if err != nil {
				return err
			}
			dispatchHostIP = config.DispatchConfig.Host
		}
	}

	if installService("postgres") {
		postgresOpts := map[string]string{
			"postgresDatabase":    config.PostgresConfig.Database,
			"postgresUser":        config.PostgresConfig.Username,
			"postgresPassword":    config.PostgresConfig.Password,
			"postgresHost":        config.PostgresConfig.Host,
			"postgresPort":        fmt.Sprintf("%d", config.PostgresConfig.Port),
			"persistence.enabled": strconv.FormatBool(config.PostgresConfig.Persistence),
		}
		mergo.Merge(&config.PostgresConfig.Chart.Opts, postgresOpts)
		err = helmInstall(out, errOut, config.PostgresConfig.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing postgres chart")
		}
	}

	if installService("api-gateway") {
		kongOpts := map[string]string{
			"services.proxyService.type":   config.APIGateway.ServiceType,
			"tls.secretName":               config.APIGateway.TLS.SecretName,
			"database":                     config.APIGateway.Database,
			"postgresql.postgresDatabase":  config.PostgresConfig.Database,
			"postgresql.postgresUser":      config.PostgresConfig.Username,
			"postgresql.postgresPassword":  config.PostgresConfig.Password,
			"postgresql.postgresHost":      config.PostgresConfig.Host,
			"postgresql.postgresPort":      fmt.Sprintf("%d", config.PostgresConfig.Port),
			"postgresql.postgresNamespace": config.PostgresConfig.Chart.Namespace,
		}
		mergo.Merge(&config.APIGateway.Chart.Opts, kongOpts)
		err = helmInstall(out, errOut, config.APIGateway.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing kong chart")
		}

		service := fmt.Sprintf("%s-kongproxy", config.APIGateway.Chart.Release)
		if config.APIGateway.ServiceType == "NodePort" {

			dispatchConfig.APIHTTPSPort, err = getK8sServiceNodePort(service, config.APIGateway.Chart.Namespace, true)
			if err != nil {
				return err
			}
			dispatchConfig.APIHTTPPort, err = getK8sServiceNodePort(service, config.APIGateway.Chart.Namespace, false)
			if err != nil {
				return err
			}

		} else if config.Ingress.ServiceType == "LoadBalancer" {
			config.APIGateway.Host, err = getK8sServiceLoadBalancerIPorHost(service, config.APIGateway.Chart.Namespace)
			if err != nil {
				return err
			}
			apiGatewayHostIP = config.APIGateway.Host
		}
	}

	if installService("openfaas") {
		openFaasOpts := map[string]string{
			"exposeServices": strconv.FormatBool(config.OpenFaas.ExposeService)}
		mergo.Merge(&config.OpenFaas.Chart.Opts, openFaasOpts)
		err = helmInstall(out, errOut, config.OpenFaas.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing openfaas chart")
		}
	}

	if installService("kubeless") {
		err = helmInstall(out, errOut, config.Kubeless.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing kubeless chart")
		}
	}

	if installService("kafka") {
		err = helmInstall(out, errOut, config.Kafka.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing kafka chart")
		}
	}

	if installService("rabbitmq") {
		rabbitMQOpts := map[string]string{
			"rabbitmqUsername":    config.RabbitMQ.Username,
			"rabbitmqPassword":    config.RabbitMQ.Password,
			"rabbitmqNodePort":    fmt.Sprintf("%d", config.RabbitMQ.Port),
			"persistence.enabled": strconv.FormatBool(config.RabbitMQ.Persistence),
		}
		mergo.Merge(&config.RabbitMQ.Chart.Opts, rabbitMQOpts)
		err = helmInstall(out, errOut, config.RabbitMQ.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing rabbitmq chart")
		}
	}

	if installService("jaeger") {
		err = helmInstall(out, errOut, config.Jaeger.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing jaeger chart")
		}
	}

	if installService("riff") {
		riffOpts := map[string]string{
			"kafka.broker.nodes":    strings.Join(config.Kafka.Brokers, ","),
			"kafka.zookeeper.nodes": strings.Join(config.Kafka.ZookeeperNodes, ","),
		}
		mergo.Merge(&config.Riff.Chart.Opts, riffOpts)
		err = helmInstall(out, errOut, config.Riff.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing riff chart")
		}
	}

	if config.DispatchConfig.ImageRegistry == nil && installService("docker-registry") {
		if config.DockerRegistry == nil {
			return errors.New("Missing docker-registry chart configuration")
		}
		err = helmInstall(out, errOut, config.DockerRegistry.Chart)
		if err != nil {
			return errors.Wrapf(err, "Error installing docker-registry chart")
		}
		serviceName := fmt.Sprintf("%s", config.DockerRegistry.Chart.Chart)
		serviceIP, err := getK8sServiceClusterIP(serviceName, config.DockerRegistry.Chart.Namespace)
		if err != nil {
			return err
		}
		registryName := fmt.Sprintf("%s:5000", serviceIP)
		config.DispatchConfig.ImageRegistry = &imageRegistryConfig{
			Name:     registryName,
			Username: "",
			Password: "",
			Email:    "",
			Insecure: true,
		}
	}

	if config.DispatchConfig.ImagePullSecret != "" {
		if config.DispatchConfig.Faas != "openfaas" {
			return errors.Errorf(
				"Must use openfaas with private registries %s is unsupported", config.DispatchConfig.Faas)
		}
		// the installed openfaas version expects the "dockercfg" credential format, but the current
		// kubectl (>1.9.1) creates "dockerconfigjson" credentials which openfaas doesn't like.  This
		// has been fixed in recent openfaas, and then the following code will go away.
		auth := fmt.Sprintf(
			"%s:%s", config.DispatchConfig.ImageRegistry.Username, config.DispatchConfig.ImageRegistry.Password)
		dockercfg := map[string]map[string]string{
			config.DispatchConfig.ImageRegistry.Name: map[string]string{
				"username": config.DispatchConfig.ImageRegistry.Username,
				"password": config.DispatchConfig.ImageRegistry.Password,
				"email":    config.DispatchConfig.ImageRegistry.Email,
				"auth":     base64.StdEncoding.EncodeToString([]byte(auth)),
			},
		}
		cfg, _ := json.Marshal(dockercfg)
		kubectlOut, err := execRunner.Run(
			"kubectl", "delete", "secret", config.DispatchConfig.ImagePullSecret,
			"-n", config.OpenFaas.Chart.Namespace)
		if err != nil {
			if !strings.Contains(string(kubectlOut), "NotFound") {
				return errors.Wrapf(err, "failed to delete existing image pull secret: %s", kubectlOut)
			}
		}
		kubectlOut, err = execRunner.Run(
			"kubectl", "create", "secret", "generic", config.DispatchConfig.ImagePullSecret,
			"-n", config.OpenFaas.Chart.Namespace,
			"--type", "kubernetes.io/dockercfg",
			"--from-literal", fmt.Sprintf(".dockercfg=%s", cfg))
		if err != nil {
			return errors.Wrapf(err, "failed to create image pull secret: %s", kubectlOut)
		}
	}

	if installService("dispatch") {
		chart := path.Join(installChartsDir, "dispatch")
		if installChartsDir != "dispatch" {
			err = helmDepUp(out, errOut, chart)
			if err != nil {
				return errors.Wrap(err, "error updating chart dependencies")
			}
		}

		// Resets the cookie every deployment if not specified
		if config.DispatchConfig.OAuth2Proxy.CookieSecret == "" {
			cookie := make([]byte, 16)
			_, err := rand.Read(cookie)
			if err != nil {
				return errors.Wrap(err, "error creating cookie secret")
			}
			config.DispatchConfig.OAuth2Proxy.CookieSecret = base64.StdEncoding.EncodeToString(cookie)
		}

		if config.DispatchConfig.OAuth2Proxy.Provider == "oidc" && config.DispatchConfig.OAuth2Proxy.OIDCIssuerURL == "" {
			return errors.New("missing oauth2Proxy.OIDCIssuerURL when the provider is specified as oidc")
		}

		// To handle the case if only dispatch service was installed.
		if config.DispatchConfig.ImageRegistry == nil {
			return errors.New("missing Image Registry configuration")
		}
		// we need to marshal username, password and email to ensure they are properly escaped
		dockerAuth := struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
		}{
			Username: config.DispatchConfig.ImageRegistry.Username,
			Password: config.DispatchConfig.ImageRegistry.Password,
			Email:    config.DispatchConfig.ImageRegistry.Email,
		}

		dockerAuthJSON, err := json.Marshal(&dockerAuth)
		if err != nil {
			return errors.Wrap(err, "error when parsing docker credentials")
		}

		dockerAuthEncoded := base64.StdEncoding.EncodeToString(dockerAuthJSON)
		apiGatewayURL := fmt.Sprintf("http://%s-kongadmin.%s:8001", config.APIGateway.Chart.Release, config.APIGateway.Chart.Namespace)
		openfaasGatewayURL := fmt.Sprintf("http://gateway.%s:8080/", config.OpenFaas.Chart.Namespace)
		rabbitMQHost := fmt.Sprintf("%s.%s", config.RabbitMQ.Chart.Release, config.RabbitMQ.Chart.Namespace)
		if config.RabbitMQ.Host != "" {
			rabbitMQHost = config.RabbitMQ.Host
		}
		jaegerAgent := config.Jaeger.Agent
		if config.Jaeger.Enabled {
			jaegerAgent = fmt.Sprintf("jaeger-agent.%s:6831", config.Jaeger.Chart.Namespace)
		}
		dispatchOpts := map[string]string{
			"global.host":                                         dispatchHost,
			"global.host_ip":                                      dispatchHostIP,
			"global.port":                                         strconv.Itoa(config.DispatchConfig.Port),
			"global.skipAuth":                                     strconv.FormatBool(config.DispatchConfig.SkipAuth),
			"global.debug":                                        strconv.FormatBool(config.DispatchConfig.Debug),
			"global.trace":                                        strconv.FormatBool(config.DispatchConfig.Trace),
			"global.data.persist":                                 strconv.FormatBool(config.DispatchConfig.PersistData),
			"global.registry.auth":                                dockerAuthEncoded,
			"global.registry.uri":                                 config.DispatchConfig.ImageRegistry.Name,
			"global.registry.insecure":                            strconv.FormatBool(config.DispatchConfig.ImageRegistry.Insecure),
			"global.tls.secretName":                               config.DispatchConfig.TLS.SecretName,
			"identity-manager.oauth2proxy.provider":               config.DispatchConfig.OAuth2Proxy.Provider,
			"identity-manager.oauth2proxy.oidcIssuerURL":          config.DispatchConfig.OAuth2Proxy.OIDCIssuerURL,
			"identity-manager.oauth2proxy.clientID":               config.DispatchConfig.OAuth2Proxy.ClientID,
			"identity-manager.oauth2proxy.clientSecret":           config.DispatchConfig.OAuth2Proxy.ClientSecret,
			"identity-manager.oauth2proxy.cookieSecret":           config.DispatchConfig.OAuth2Proxy.CookieSecret,
			"identity-manager.ingress.host":                       dispatchHost,
			"global.db.backend":                                   config.DispatchConfig.Database,
			"global.db.host":                                      config.PostgresConfig.Host,
			"global.db.port":                                      fmt.Sprintf("%d", config.PostgresConfig.Port),
			"global.db.user":                                      config.PostgresConfig.Username,
			"global.db.password":                                  config.PostgresConfig.Password,
			"global.db.release":                                   config.PostgresConfig.Chart.Release,
			"global.db.namespace":                                 config.PostgresConfig.Chart.Namespace,
			"global.rabbitmq.username":                            config.RabbitMQ.Username,
			"global.rabbitmq.password":                            config.RabbitMQ.Password,
			"global.rabbitmq.host":                                rabbitMQHost,
			"global.rabbitmq.port":                                fmt.Sprintf("%d", config.RabbitMQ.Port),
			"global.kafka.brokers":                                fmt.Sprintf("{%s}", strings.Join(config.Kafka.Brokers, ",")),
			"global.tracer.endpoint":                              jaegerAgent,
			"api-manager.gateway.host":                            apiGatewayURL,
			"function-manager.faas.selected":                      config.DispatchConfig.Faas,
			"function-manager.faas.imagePullSecret":               config.DispatchConfig.ImagePullSecret,
			"function-manager.faas.openfaas.gateway":              openfaasGatewayURL,
			"function-manager.faas.openfaas.namespace":            config.OpenFaas.Chart.Namespace,
			"function-manager.faas.riff.namespace":                config.Riff.Chart.Namespace,
			"event-manager.transport":                             config.DispatchConfig.EventTransport,
			"service-manager.catalog.selected":                    config.DispatchConfig.Service.Catalog,
			"service-manager.catalog.k8sservicecatalog.namespace": config.DispatchConfig.Service.K8sServiceCatalog.Namespace,
		}

		// If unset values default to chart values
		if config.DispatchConfig != nil && config.DispatchConfig.Image != nil {
			if config.DispatchConfig.Image.Host != "" {
				dispatchOpts["global.image.host"] = config.DispatchConfig.Image.Host
			}
			if config.DispatchConfig.Image.Tag != "" {
				dispatchOpts["global.image.tag"] = config.DispatchConfig.Image.Tag
			}
		}
		if installDebug {
			for k, v := range dispatchOpts {
				fmt.Fprintf(out, "%v: %v\n", k, v)
			}
		}
		mergo.Merge(&config.DispatchConfig.Chart.Opts, dispatchOpts)
		err = helmInstall(out, errOut, config.DispatchConfig.Chart)
		if err != nil {
			return errors.Wrapf(err, "error installing dispatch chart")
		}
		if installService("certs") && !installDryRun && config.DispatchConfig.TLS.Ca == "letsEncrypt" {
			// Technically we have the route53 secrets and could configure
			// a job as part of the certificate chart to update the records
			// this is future work
			fmt.Print("##############################\n")
			fmt.Print("Add the following DNS records:\n")
			fmt.Printf("\t%s\t\t%s\n", apiGatewayHostIP, apiGatewayHost)
			fmt.Printf("\t%s\t\t%s\n", dispatchHostIP, dispatchHost)
			fmt.Print("##############################\n")
		}
	}
	err = writeConfig(out, errOut, configDir, config)
	return err
}
