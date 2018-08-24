/*
Copyright 2016 The Kubernetes Authors.

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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_ClusterServiceBrokerSpec(spec *ClusterServiceBrokerSpec) {
	setCommonServiceBrokerDefaults(&spec.CommonServiceBrokerSpec)
}

func SetDefaults_ServiceBrokerSpec(spec *ServiceBrokerSpec) {
	setCommonServiceBrokerDefaults(&spec.CommonServiceBrokerSpec)
}

func setCommonServiceBrokerDefaults(spec *CommonServiceBrokerSpec) {
	if spec.RelistBehavior == "" {
		spec.RelistBehavior = ServiceBrokerRelistBehaviorDuration
	}
}

func SetDefaults_ServiceBinding(binding *ServiceBinding) {
	// If not specified, make the SecretName default to the binding name
	if binding.Spec.SecretName == "" {
		binding.Spec.SecretName = binding.Name
	}
}
