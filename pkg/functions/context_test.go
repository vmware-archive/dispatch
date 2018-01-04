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
	ctx.ReadLogs(bytes.NewReader([]byte("foo\nbar\n")))
	assert.Equal(t, []string{"foo", "bar"}, ctx.Logs())
}

func TestContext_LogsBS(t *testing.T) {
	ctx := Context{}
	ctx["logs"] = &struct {
		bs string
	}{}
	assert.Len(t, ctx.Logs(), 0)
	ctx["logs"] = []interface{}{"ln1", "ln2"}
	assert.Equal(t, []string{"ln1", "ln2"}, ctx.Logs())
}

func TestContext_LogsNil(t *testing.T) {
	ctx := Context{}
	assert.Len(t, ctx.Logs(), 0)
}
