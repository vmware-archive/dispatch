///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi"
	"github.com/vmware/dispatch/pkg/secret-store/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secret-store/service"
	"github.com/vmware/dispatch/pkg/secret-store/web"
)

type secretsConfig struct {
	K8sConfig    string `mapstructure:"kubeconfig" json:"kubeconfig,omitempty,omitempty"`
	K8sNamespace string `mapstructure:"namespace" json:"namespace,omitempty,omitempty"`
}

// NewCmdSecrets creates a subcommand to run secret store
func NewCmdSecrets(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "secret-store",
		Short:  i18n.T("Run Dispatch Secret Store"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Secrets),
		Run: func(cmd *cobra.Command, args []string) {
			runSecrets(config)
		},
	}
	cmd.SetOutput(out)

	cmd.Flags().String("kubeconfig", "", "Path to kubernetes config file")
	cmd.Flags().String("namespace", "default", "Kubernetes namespace")
	return cmd
}

func runSecrets(config *serverConfig) {
	store := entityStore(config)

	var k8sConfig *rest.Config
	var err error
	if config.Secrets.K8sConfig == "" {
		k8sConfig, err = rest.InClusterConfig()
	} else {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", config.Secrets.K8sConfig)
	}
	if err != nil {
		log.Fatalf("Error getting kubernetes config: %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %+v", err)
	}

	secretsService := &service.K8sSecretsService{
		EntityStore: store,
		SecretsAPI:  clientset.CoreV1().Secrets(config.Secrets.K8sNamespace),
	}

	secretsHandler := initSecrets(config, secretsService)

	handler := addMiddleware(secretsHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initSecrets(config *serverConfig, secretsService service.SecretsService) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewSecretStoreAPI(swaggerSpec)

	handlers := web.NewHandlers(secretsService)

	web.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
