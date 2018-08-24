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
	"fmt"

	"k8s.io/apimachinery/pkg/version"
)

// ServerVersion asks the Service Catalog API Server for the version.Info object and returns it.
func (sdk *SDK) ServerVersion() (*version.Info, error) {
	serverVersion, err := sdk.ServiceCatalogClient.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("unable to get version, %v", err)
	}

	return serverVersion, nil
}
