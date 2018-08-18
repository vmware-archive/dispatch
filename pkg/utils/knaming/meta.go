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

func ToJSONString(obj interface{}) string {
	bs, err := json.Marshal(obj)
	if err != nil {
		// TODO the right thing
		panic(errors.Wrapf(err, "could not JSON-marshal this: %+v", obj))
	}
	return string(bs)
}

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

func ToLabelSelector(y map[string]string) string {
	var ss []string
	for k, v := range y {
		ss = append(ss, k+"="+v)
	}
	return strings.Join(ss, ",")
}

func FunctionName(project, name string) string {
	return "d-fn-" + project + "-" + name
}

func AdjustMeta(meta *dapi.Meta, org string, project string) {
	if meta.Org == DefaultOrg {
		meta.Org = org
	}
	if meta.Project == DefaultProject {
		meta.Project = project
	}
}
