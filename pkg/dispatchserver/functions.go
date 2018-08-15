///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/function-manager"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/kubeless"
	"github.com/vmware/dispatch/pkg/functions/noop"
	"github.com/vmware/dispatch/pkg/functions/openfaas"
	"github.com/vmware/dispatch/pkg/functions/riff"
)

type functionsConfig struct {
	FaaS                string                       `mapstructure:"faas" json:"faas,omitempty"`
	ImagePullSecret     string                       `mapstructure:"image-pull-secret" json:"image-pull-secret,omitempty"`
	K8sConfig           string                       `mapstructure:"kubeconfig" json:"kubeconfig,omitempty"`
	FuncDefaultLimits   *functions.FunctionResources `mapstructure:"func-default-limits" json:"func-default-limits,omitempty"`
	FuncDefaultRequests *functions.FunctionResources `mapstructure:"func-default-requests" json:"func-default-requests,omitempty"`
	OpenFaaSNamespace   string                       `mapstructure:"openfaas-namespace" json:"openfaas-namespace,omitempty"`
	OpenFaaSGateway     string                       `mapstructure:"openfaas-gateway" json:"openfaas-gateway,omitempty"`
	RiffKafkaBrokers    []string                     `mapstructure:"riff-kafka-brokers" json:"riff-kafka-brokers,omitempty"`
	RiffNamespace       string                       `mapstructure:"riff-namespace" json:"riff-namespace,omitempty"`
	KubelessNamespace   string                       `mapstructure:"kubeless-namespace" json:"kubeless-namespace,omitempty"`
	FileImageManager    string                       `mapstructure:"file-image-manager" json:"file-image-manager,omitempty"`
}

func faasDriver(config functionsConfig, zk string) functions.FaaSDriver {
	var faas functions.FaaSDriver
	var err error
	switch config.FaaS {
	case "openfaas":
		faas, err = openfaas.New(&openfaas.Config{
			Gateway:             config.OpenFaaSGateway,
			K8sConfig:           config.K8sConfig,
			FuncNamespace:       config.OpenFaaSNamespace,
			FuncDefaultRequests: config.FuncDefaultRequests,
			FuncDefaultLimits:   config.FuncDefaultLimits,
			ImagePullSecret:     config.ImagePullSecret,
		})
	case "riff":
		faas, err = riff.New(&riff.Config{
			KafkaBrokers:        config.RiffKafkaBrokers,
			K8sConfig:           config.K8sConfig,
			FuncNamespace:       config.RiffNamespace,
			FuncDefaultRequests: config.FuncDefaultRequests,
			FuncDefaultLimits:   config.FuncDefaultLimits,
			ZookeeperLocation:   zk,
		})
	case "kubeless":
		faas, err = kubeless.New(&kubeless.Config{
			K8sConfig:       config.K8sConfig,
			FuncNamespace:   config.KubelessNamespace,
			ImagePullSecret: config.ImagePullSecret,
		})
	case "noop":
		faas, err = noop.New(&noop.Config{})
	default:
		err = fmt.Errorf("FaaS %s not supported", config.FaaS)
	}
	if err != nil {
		log.Fatalf("Error starting %s driver: %+v", config.FaaS, err)
	}
	return faas
}

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

// NewCmdFunctions creates a subcommand to create functions manager
func NewCmdFunctions(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "function-manager",
		Short:  i18n.T("Run Dispatch Functions Manager"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Functions),
		Run: func(cmd *cobra.Command, args []string) {
			runFunctions(config)
		},
	}

	cmd.Flags().String("faas", "openfaas", "FaaS backend to use (openfaas|kubeless|riff|noop)")
	cmd.Flags().String("image-pull-secret", "", "Base64-encoded docker secrets used when pulling images")
	cmd.Flags().String("kubeconfig", "", "Path to kubeconfig")
	cmd.Flags().String("openfaas-namespace", "", "Namespace to use when deploying openfaas functions")
	cmd.Flags().String("openfaas-gateway", "", "OpenFaas gateway URL")
	cmd.Flags().String("riff-namespace", "", "Namespace to use when deploying riff functions")
	cmd.Flags().StringSlice("riff-kafka-brokers", []string{}, "Kafka brokers to use when communicating with Riff")
	cmd.Flags().String("kubeless-namespace", "", "Namespace to use when deploying Kubeless functions")
	cmd.Flags().String("file-image-manager", "", "Path to file image manager, useful for testing")

	cmd.SetOutput(out)
	return cmd
}

func runFunctions(config *serverConfig) {

	fnHandler, shutdown := initFunctions(config)
	defer shutdown()

	handler := addMiddleware(fnHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

func initFunctions(config *serverConfig) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionManagerAPI(swaggerSpec)
	handlers := functionmanager.NewHandlers(config.Functions.K8sConfig)
	functionmanager.ConfigureHandlers(api, handlers)

	return api.Serve(nil), func() {}
}
