///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package knaming

import (
	"testing"

	"encoding/json"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api/v1"
)

func TestGetName(t *testing.T) {
	fnName := "hello-py"
	fn := &v1.Function{
		Name: &fnName,
		Meta: v1.Meta{
			Project: "test-project",
			Org:     "test-organization",
		},
	}
	meta := ToObjectMeta(fn)
	assert.Equal(t, "d-function-test-project-hello-py", meta.Name)
}

func TestToLabelSelector(t *testing.T) {
	data := map[string]string{
		"a": "97",
		"b": "98",
	}
	labels := ToLabelSelector(data)
	assert.Equal(t, "a=97,b=98", labels)
}

func TestToObjectMeta(t *testing.T) {
	fnName := "hello-py"
	dockerURL := "dockerURL"
	fn := &v1.Function{
		Name: &fnName,
		Meta: v1.Meta{
			Project: "test-project",
			Org:     "test-organization",
			Name:    "d-function-test-project-hello-py",
		},
		Image: &dockerURL,
	}
	objectMeta := ToObjectMeta(fn)
	assert.Equal(t, *fn.Name, objectMeta.Labels[NameLabel])
	assert.Equal(t, fn.Meta.Project, objectMeta.Labels[ProjectLabel])
	assert.Equal(t, fn.Meta.Org, objectMeta.Labels[OrgLabel])
	bytes, _ := json.Marshal(fn)
	fmt.Printf(string(bytes))
	assert.Equal(t, string(bytes), objectMeta.Annotations[InitialObjectAnnotation])
}

func TestFromObjectMeta(t *testing.T) {
	fnName := "hello-py"
	dockerURL := "dockerURL"
	fn := &v1.Function{
		Name: &fnName,
		Meta: v1.Meta{
			Project: "test-project",
			Org:     "test-organization",
			Name:    "d-function-test-project-hello-py",
		},
		Image: &dockerURL,
	}
	objectMeta := ToObjectMeta(fn)
	reboundFn := &v1.Function{}
	FromObjectMeta(objectMeta, reboundFn)
	assert.Equal(t, fn, reboundFn)
}
