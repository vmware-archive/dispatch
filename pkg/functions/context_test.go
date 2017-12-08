///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_Logs(t *testing.T) {
	ctx := Context{}
	ctx.SetLogs(bytes.NewReader([]byte("foo\nbar\n")))
	assert.Equal(t, []string{"foo", "bar"}, ctx.Logs())
}

func TestContext_LogsBS(t *testing.T) {
	ctx := Context{}
	ctx["logs"] = &struct {
		bs string
	}{}
	assert.Len(t, ctx.Logs(), 0)
}

func TestContext_LogsNil(t *testing.T) {
	ctx := Context{}
	assert.Len(t, ctx.Logs(), 0)
}
