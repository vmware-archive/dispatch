///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package dispatchserver_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/dispatchserver"
)

func TestCmdMainCommand(t *testing.T) {
	var buf bytes.Buffer

	cli := dispatchserver.NewCLI(&buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{})
	err := cli.Execute()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "Dispatch is a batteries-included serverless framework."))
}
