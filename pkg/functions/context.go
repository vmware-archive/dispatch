///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"bufio"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/trace"
)

func (ctx Context) Logs() []string {
	defer trace.Tracef("")()

	log.Debugf(`Logs from ctx["logs"]: %v`, ctx["logs"])
	logs, _ := ctx["logs"].([]string)
	return logs
}

func (ctx Context) SetLogs(reader io.Reader) {
	defer trace.Tracef("")()

	ctx["logs"] = getLogs(reader)
}

func getLogs(reader io.Reader) []string {
	defer trace.Tracef("")()

	scanner := bufio.NewScanner(reader)
	var logs []string
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	return logs
}
