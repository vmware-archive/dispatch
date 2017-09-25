///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdCreate(t *testing.T) {
	var buf bytes.Buffer
	path := createConfig(t, "")
	defer os.Remove(path) // clean up

	cli := NewVSCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{fmt.Sprintf("--config=%s", path), "create"})
	err := cli.Execute()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "Create a resource"))
}
