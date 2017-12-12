///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckErr(t *testing.T) {
	err := errors.New("test error")
	var called bool
	var testFun = func(msg string, code int) {
		assert.Equal(t, "test error", msg)
		called = true
	}
	checkErr(err, testFun)
	assert.True(t, called)
}
