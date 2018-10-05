///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/utils"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/dispatch/pkg/secrets/gen/restapi"
	"github.com/vmware/dispatch/pkg/secrets/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/secrets/service"
	"github.com/vmware/dispatch/pkg/secrets/web"
)

func initSecrets(config *serverConfig) http.Handler {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewSecretsAPI(swaggerSpec)

	// Need to refactor some of the knative helpers out of functions to reuse
	// across Dispatch
	k8sConfig, err := utils.KubeClientConfig(config.K8sConfig)
	if err != nil {
		log.Fatalf("Error getting kubernetes config: %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %+v", err)
	}

	secretsService := &service.K8sSecretsService{
		K8sAPI: clientset.CoreV1(),
	}

	handlers := web.NewHandlers(secretsService)

	web.ConfigureHandlers(api, handlers)

	return api.Serve(nil)
}
