/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package servicecatalog

import (
	svcatclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// SDK wrapper around the generated Go client for the Kubernetes Service Catalog
type SDK struct {
	K8sClient            kubernetes.Interface
	ServiceCatalogClient svcatclient.Interface
}

// ServiceCatalog is the underlying generated Service Catalog versioned interface
// It should be used instead of accessing the client directly.
func (sdk *SDK) ServiceCatalog() v1beta1.ServicecatalogV1beta1Interface {
	return sdk.ServiceCatalogClient.ServicecatalogV1beta1()
}

// Core is the underlying generated Core API versioned interface
// It should be used instead of accessing the client directly.
func (sdk *SDK) Core() corev1.CoreV1Interface {
	return sdk.K8sClient.CoreV1()
}
