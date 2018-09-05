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

var secrets = v1.SecretValue{
	"foo": "bar",
	"baz": "qux",
}

func f1() *v1.Function {
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

func s1() *v1.Secret {
	return &v1.Secret{
		Meta: v1.Meta{
			Org:     testOrg,
			Project: testProject,
			Name:    secret1,
		},
		Secrets: secrets,
	}
}

func TestToObjectMeta_Function(t *testing.T) {
	f1 := *f1()

	f1om := ToObjectMeta(f1.Meta, f1)
	var f1Parsed v1.Function
	FromJSONString(f1om.Annotations[InitialObjectAnnotation], &f1Parsed)
	assert.Equal(t, f1, f1Parsed)
}

func TestToObjectMeta_Secret(t *testing.T) {

	s1 := *s1()
	assert.Equal(t, secrets, s1.Secrets)

	s1om := ToObjectMeta(s1.Meta, s1)
	var s1Parsed v1.Secret
	FromJSONString(s1om.Annotations[InitialObjectAnnotation], &s1Parsed)
	assert.Nil(t, s1Parsed.Secrets)

	s1.Secrets = nil
	assert.Equal(t, s1, s1Parsed)
}
