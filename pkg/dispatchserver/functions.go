///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
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
)

type functionsConfig struct {
	K8sConfig string `mapstructure:"kubeconfig" json:"kubeconfig,omitempty"`
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

	cmd.Flags().String("kubeconfig", "", "Path to kubeconfig")

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
