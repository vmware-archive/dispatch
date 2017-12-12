///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package functions

import (
	"bufio"
	"io"
)

func (ctx Context) Logs() []string {
	return ctx["logs"].([]string)
}

func (ctx Context) SetLogs(reader io.Reader) {
	ctx["logs"] = getLogs(reader)
}

func getLogs(reader io.Reader) []string {
	scanner := bufio.NewScanner(reader)
	var logs []string
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	return logs
}
