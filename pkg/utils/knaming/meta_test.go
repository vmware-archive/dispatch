///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package knaming

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api/v1"
)

const (
	testOrg     = "vmware"
	testProject = "dispatch"

	fn1 = "f1"

	secret1 = "secret1"
	secret2 = "secret2"
)

func createFunc() *v1.Function {
	return &v1.Function{
		Meta: v1.Meta{
			Org:     testOrg,
			Project: testProject,
			Name:    fn1,
		},
		FunctionImageURL: "dispatchframework/test-fn1",
		Secrets:          []string{secret1, secret2},
	}
}

func TestToObjectMeta_Function(t *testing.T) {
	f1 := *createFunc()

	f1om := ToObjectMeta(&f1)
	var f1Parsed v1.Function
	FromObjectMeta(f1om, &f1Parsed)
	assert.Equal(t, f1, f1Parsed)
}

func TestGetName(t *testing.T) {
	fnName := "hello-py"
	fn := &v1.Function{
		Name: &fnName,
		Meta: v1.Meta{
			Project: "test-project",
			Org:     "test-organization",
			Name:    fnName,
		},
	}
	meta := ToObjectMeta(fn)
	assert.Equal(t, "d-fn-test-project-hello-py", meta.Name)
}

func TestToLabelSelector(t *testing.T) {
	data := map[string]string{
		"a": "97",
		"b": "98",
	}
	labels := ToLabelSelector(data)
	assert.Equal(t, "a=97,b=98", labels)
}
