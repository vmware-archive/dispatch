///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package knaming

import (
	"encoding/json"
	"reflect"
	"strings"

	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Constants for working with Knative
const (
	NameLabel    = "dispatchframework.io/name"
	ProjectLabel = "dispatchframework.io/project"
	OrgLabel     = "dispatchframework.io/org"

	KnTypeLabel = "dispatchframework.io/type"

	TheSecretKey = "key"

	InitialObjectAnnotation = "dispatchframework.io/initialObject"
)

// ToJSONString JSON-encodes a Dispatch API object
func ToJSONString(obj interface{}) string {
	bs, err := json.Marshal(obj)
	if err != nil {
		// TODO the right thing
		panic(errors.Wrapf(err, "could not JSON-marshal object: %+v", obj))
	}
	return string(bs)
}

// FromObjectMeta decodes a Dispatch API object from K8S Object Meta
func FromObjectMeta(objectMeta *v1.ObjectMeta, obj interface{}) error {
	jsonString := objectMeta.Annotations[InitialObjectAnnotation]
	if err := json.Unmarshal([]byte(jsonString), obj); err != nil {
		log.Debugf("could not unmarshal from JSON: '%s'", jsonString)
		return errors.Wrapf(err, "could not unmarshal from JSON: '%s'", jsonString)
	}
	return nil
}

// ToObjectMeta produces a k8s API *ObjectMeta from the original Dispatch object
func ToObjectMeta(initialObject interface{}) *v1.ObjectMeta {
	t := reflect.TypeOf(initialObject)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return nil
	}

	objType := t.Elem().Name()

	// TODO: handle other error conditions
	v := reflect.ValueOf(initialObject).Elem()

	dispatchName := v.FieldByName("Name").Elem().String()

	metaField := v.FieldByName("Meta")
	dispatchProject := metaField.FieldByName("Project").String()
	dispatchOrg := metaField.FieldByName("Org").String()
	k8sName := metaField.FieldByName("Name").String()

	labels := map[string]string{
		NameLabel:    dispatchName,
		ProjectLabel: dispatchProject,
		OrgLabel:     dispatchOrg,
		KnTypeLabel:  objType,
	}

	return &v1.ObjectMeta{
		Name:        k8sName,
		Labels:      labels,
		Annotations: map[string]string{InitialObjectAnnotation: ToJSONString(initialObject)},
	}
}

// ToLabelSelector produces a k8s API label selector string
func ToLabelSelector(y map[string]string) string {
	labelSelector := v1.LabelSelector{
		MatchLabels: y,
	}
	return v1.FormatLabelSelector(&labelSelector)
}

// GetName returns k8s API name
func GetK8SName(objType, project, name string) string {
	objType = strings.ToLower(objType)
	return fmt.Sprintf("d-%s-%s-%s", objType, project, name)
}

//SecretEnvVarName returns the env var name to hold the secret
func SecretEnvVarName(name string) string {
	return "d_secret_" + name
}

//SecretName returns k8s API name of the Dispatch secret
func SecretName(meta dapi.Meta) string {
	return "d-secret-" + meta.Project + "_" + meta.Name
}
