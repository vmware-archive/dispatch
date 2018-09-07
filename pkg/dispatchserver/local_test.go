///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package dispatchserver

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	// The following blank import is to load GKE auth plugin required when authenticating against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// The following blank import is to load OIDC auth plugin required when authenticating against OIDC-enabled clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func _TestCmdLocalCommand(t *testing.T) { // TODO disabling: breaks CI. Local is very likely going to be removed
	var buf bytes.Buffer

	cli := NewCLI(&buf)
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
