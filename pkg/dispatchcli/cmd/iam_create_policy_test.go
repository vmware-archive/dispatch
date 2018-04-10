///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdIamCreatePolicy(t *testing.T) {
	var buf bytes.Buffer

	cli := NewCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{"iam", "create", "policy", "--help"})
	err := cli.Execute()

	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "Create a dispatch policy"))
}
