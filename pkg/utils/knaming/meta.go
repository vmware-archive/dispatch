///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package knaming

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
)

// Constants for working with Knative
const (
	NameLabel    = "dispatchframework.io/name"
	ProjectLabel = "dispatchframework.io/project"
	OrgLabel     = "dispatchframework.io/org"

	KnTypeLabel = "knative.dev/type"

	TheSecretKey = "secret"

	InitialObjectAnnotation = "dispatchframework.io/initialObject"
)

var typeNameMap = map[string]string{
	"Function":        "fn",
	"Secret":          "secret",
	"EventDriverType": "dt",
}

// ToJSONString JSON-encodes a Dispatch API object
func ToJSONString(obj interface{}) string {
	return string(ToJSONBytes(obj))
}

// ToJSONBytes JSON-encodes a Dispatch API object
func ToJSONBytes(obj interface{}) []byte {
	bs, err := json.Marshal(obj)
	if err != nil {
		// TODO the right thing
		panic(errors.Wrapf(err, "could not JSON-marshal object: %+v", obj))
	}
	return bs
}

// FromJSONBytes decodes a Dispatch API object from JSON string
func FromJSONBytes(bs []byte, obj interface{}) error {
	if err := json.Unmarshal(bs, obj); err != nil {
		return errors.Wrapf(err, "could not unmarshal from JSON: '%s'", bs)
	}
	return nil
}

// FromObjectMeta decodes a Dispatch API object from K8S Object Meta
func FromObjectMeta(objectMeta *v1.ObjectMeta, obj interface{}) error {
	jsonString := objectMeta.Annotations[InitialObjectAnnotation]
	return FromJSONBytes([]byte(jsonString), obj)
}

// ToObjectMeta produces a k8s API *ObjectMeta from the original Dispatch object
func ToObjectMeta(initialObject interface{}) *v1.ObjectMeta {
	t := reflect.TypeOf(initialObject)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		// TODO: return an err
		return nil
	}

	objType := t.Elem().Name()

	// TODO: handle other error conditions
	v := reflect.ValueOf(initialObject).Elem()

	metaField := v.FieldByName("Meta")
	dispatchMeta := dapi.Meta{
		Project: metaField.FieldByName("Project").String(),
		Org:     metaField.FieldByName("Org").String(),
		Name:    metaField.FieldByName("Name").String(),
	}

	labels := map[string]string{
		NameLabel:    dispatchMeta.Name,
		ProjectLabel: dispatchMeta.Project,
		OrgLabel:     dispatchMeta.Org,
		KnTypeLabel:  objType,
	}

	annotations := map[string]string{InitialObjectAnnotation: ToJSONString(initialObject)}

	return &v1.ObjectMeta{
		Name:        GetKnName(objType, dispatchMeta),
		Labels:      labels,
		Annotations: annotations,
	}
}

// ToLabelSelector produces a k8s API label selector string
func ToLabelSelector(y map[string]string) string {
	labelSelector := v1.LabelSelector{
		MatchLabels: y,
	}
	return v1.FormatLabelSelector(&labelSelector)
}

// GetKnName returns k8s API name
func GetKnName(objType string, meta dapi.Meta) string {
	objName := typeNameMap[objType]
	return fmt.Sprintf("d-%s-%s-%s", objName, meta.Project, meta.Name)
}

// SecretEnvVarName returns the env var name to hold the secret
func SecretEnvVarName(name string) string {
	return "d_secret_" + name
}

// SecretName returns k8s API name of the Dispatch secret
func SecretName(meta dapi.Meta) string {
	return GetKnName("Secret", meta)
}
