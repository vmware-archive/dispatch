///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCLI(t *testing.T) {
	var buf bytes.Buffer

	cli := NewVSCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{})
	err := cli.Execute()
	assert.Nil(t, err)
}
