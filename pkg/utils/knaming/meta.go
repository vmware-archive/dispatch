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

	DefaultOrg     = "default"
	DefaultProject = "default"

	InitialObjectAnnotation = "dispatchframework.io/initialObject"
)

//ToJSONString JSON-encodes a Dispatch API object
func ToJSONString(obj interface{}) string {
	bs, err := json.Marshal(obj)
	if err != nil {
		// TODO the right thing
		panic(errors.Wrapf(err, "could not JSON-marshal object: %+v", obj))
	}
	return string(bs)
}

//FromJSONString decodes a Dispatch API object from JSON string
func FromJSONString(jsonString string, obj interface{}) error {
	if err := json.Unmarshal([]byte(jsonString), obj); err != nil {
		return errors.Wrapf(err, "could not unmarshal from JSON: '%s'", jsonString)
	}
	return nil
}

//ToObjectMeta produces a k8s API *ObjectMeta from Dispatch API Meta and the original Dispatch object
func ToObjectMeta(meta dapi.Meta, initialObject interface{}) *v1.ObjectMeta {
	name := ""
	labels := map[string]string{NameLabel: meta.Name, ProjectLabel: meta.Project, OrgLabel: meta.Org}

	switch initialObject.(type) {
	case dapi.Function:
		name = FunctionName(meta.Project, meta.Name)
		labels[KnTypeLabel] = FunctionKnType
	}

	if name == "" {
		// TODO handle it
		panic(errors.New("name is empty"))
	}

	return &v1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: map[string]string{InitialObjectAnnotation: ToJSONString(initialObject)},
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
func FunctionName(project, name string) string {
	return "d-fn-" + project + "-" + name
}

//AdjustMeta replaces default values of Org and Project fields of Meta with specified values
func AdjustMeta(meta *dapi.Meta, org string, project string) {
	if meta.Org == DefaultOrg {
		meta.Org = org
	}
	if meta.Project == DefaultProject {
		meta.Project = project
	}
}
