///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package http

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePKI(t *testing.T) {
	privateKey, certificate, err := GeneratePKI([]string{"example.com"})
	assert.NoError(t, err)
	assert.NotEmpty(t, privateKey)
	assert.NotEmpty(t, certificate)
	if privateKey != "" {
		err = os.Remove(privateKey)
		assert.NoError(t, err)
	}
	if certificate != "" {
		err = os.Remove(certificate)
		assert.NoError(t, err)
	}
}

func TestGeneratePKIEmpty(t *testing.T) {
	privateKey, certificate, err := GeneratePKI(nil)
	assert.Error(t, err)
	assert.Empty(t, privateKey)
	assert.Empty(t, certificate)
}
