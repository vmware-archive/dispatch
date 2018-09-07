///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package dispatchserver

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdMainCommand(t *testing.T) {
	var buf bytes.Buffer

	cli := NewCLI(&buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{})
	err := cli.Execute()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "Dispatch is a batteries-included serverless framework."))
}
