///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
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
