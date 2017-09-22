///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultConfig = `host: localhost
port: 8000
organization: vmware`

func createConfig(t *testing.T, config string) string {
	if config == "" {
		config = defaultConfig
	}
	content := []byte(config)
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	tmpfn := filepath.Join(tmpDir, "vs.yaml")
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		t.Fatal(err)
	}
	return tmpfn
}

func TestRootCLI(t *testing.T) {
	var buf bytes.Buffer
	path := createConfig(t, "")
	defer os.Remove(path) // clean up

	cli := NewVSCLI(os.Stdin, &buf, &buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{fmt.Sprintf("--config=%s", path)})
	err := cli.Execute()
	assert.Nil(t, err)
}
