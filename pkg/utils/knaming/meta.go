///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package knaming

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Constants for working with Knative
const (
	NameLabel    = "dispatchframework.io/name"
	ProjectLabel = "dispatchframework.io/project"
	OrgLabel     = "dispatchframework.io/org"

	KnTypeLabel = "knative.dev/type"

	FunctionKnType = "function"

	TheSecretKey = "secret"

	InitialObjectAnnotation = "dispatchframework.io/initialObject"
)

//ToJSONString JSON-encodes a Dispatch API object
func ToJSONString(obj interface{}) string {
	return string(ToJSONBytes(obj))
}

//ToJSONBytes JSON-encodes a Dispatch API object
func ToJSONBytes(obj interface{}) []byte {
	bs, err := json.Marshal(obj)
	if err != nil {
		// TODO the right thing
		panic(errors.Wrapf(err, "could not JSON-marshal object: %+v", obj))
	}
	return bs
}

//FromJSONString decodes a Dispatch API object from JSON string
func FromJSONString(jsonString string, obj interface{}) error {
	return FromJSONBytes([]byte(jsonString), obj)
}

//FromJSONBytes decodes a Dispatch API object from JSON string
func FromJSONBytes(bs []byte, obj interface{}) error {
	if err := json.Unmarshal(bs, obj); err != nil {
		return errors.Wrapf(err, "could not unmarshal from JSON: '%s'", bs)
	}
	return nil
}

//ToObjectMeta produces a k8s *ObjectMeta from Dispatch `meta` and the `initialObject` (MUST be non-pointer)
func ToObjectMeta(meta dapi.Meta, initialObject interface{}) v1.ObjectMeta {
	name := ""
	labels := map[string]string{NameLabel: meta.Name, ProjectLabel: meta.Project, OrgLabel: meta.Org}

	switch typedObject := initialObject.(type) {
	case dapi.Function:
		name = FunctionName(meta)
		labels[KnTypeLabel] = FunctionKnType
	case dapi.Secret:
		name = SecretName(meta)
		typedObject.Secrets = nil // omit secret data
		initialObject = typedObject
	case dapi.BaseImage:
		name = BaseImageName(meta)
		initialObject = typedObject
	case dapi.Image:
		name = ImageName(meta)
		initialObject = typedObject
	case dapi.Endpoint:
		name = EndpointName(meta)
		initialObject = typedObject
	default:
		// TODO handle it
		panic(errors.New("unknown type"))
	}

	annotations := map[string]string{InitialObjectAnnotation: ToJSONString(initialObject)}

	return v1.ObjectMeta{
		Name:        name,
		Namespace:   meta.Org,
		Labels:      labels,
		Annotations: annotations,
	}
}

//ToLabelSelector produces a k8s API label selector string
func ToLabelSelector(y map[string]string) string {
	var ss []string
	for k, v := range y {
		ss = append(ss, k+"="+v)
	}
	return strings.Join(ss, ",")
}

//FunctionName returns k8s API name of the Dispatch function
func FunctionName(meta dapi.Meta) string {
	return "d-fn-" + meta.Project + "-" + meta.Name
}

//SecretEnvVarName returns the env var name to hold the secret
func SecretEnvVarName(name string) string {
	return "d_secret_" + name
}

//SecretName returns k8s API name of the Dispatch secret
func SecretName(meta dapi.Meta) string {
	return "d-secret-" + meta.Project + "-" + meta.Name
}

//BaseImageName returns k8s API name of the Dispatch baseimage
func BaseImageName(meta dapi.Meta) string {
	return "d-base-image-" + meta.Project + "-" + meta.Name
}

//ImageName returns k8s API name of the Dispatch image
func ImageName(meta dapi.Meta) string {
	return "d-image-" + meta.Project + "-" + meta.Name
}

//EndpointName returns k8s API name of the Dispatch endpoint
func EndpointName(meta dapi.Meta) string {
	return "d-endpoint-" + meta.Project + "-" + meta.Name
}
