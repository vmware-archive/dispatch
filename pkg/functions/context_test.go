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
