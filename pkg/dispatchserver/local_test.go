///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package dispatchserver_test

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/vmware/dispatch/pkg/dispatchserver"
)

func TestCmdLocalCommand(t *testing.T) {
	var buf bytes.Buffer

	cli := dispatchserver.NewCLI(&buf)
	cli.SetOutput(&buf)
	cli.SetArgs([]string{"local", "--port", "0"})
	go cli.Execute()
	ticker := time.NewTicker(time.Millisecond * 100)
	timeout := time.NewTimer(time.Second * 5)
	for {
		select {
		case <-timeout.C:
			t.Errorf("Timeout waiting for dispatch-server local to start listening. Content of the output: %s", buf.String())
			return
		case <-ticker.C:
		}
		if strings.Contains(buf.String(), "serving HTTP traffic") {
			break
		}
	}
	regex := regexp.MustCompile(`http://127\.0\.0\.1:(\d+)`)
	res := regex.FindStringSubmatch(buf.String())
	if len(res) < 2 {
		t.Errorf("Unable to find port in %s", buf.String())
		return
	}
	port := res[1]

	_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/v1/secret", port))
	if err != nil {
		t.Errorf("Error when connecting to dispatch server: %v", err)
	}
}
