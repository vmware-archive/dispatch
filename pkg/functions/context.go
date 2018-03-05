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

const logsKey = "logs"

// Logs returns the logs as a list of strings
func (ctx Context) Logs() []string {
	defer trace.Tracef("")()

	log.Debugf(`Logs from ctx["logs"]: %#v`, ctx["logs"])
	switch logs := ctx[logsKey].(type) {
	case []string:
		return logs
	case []interface{}:
		var r []string
		for _, l := range logs {
			s, ok := l.(string)
			if !ok {
				break
			}
			r = append(r, s)
		}
		return r
	}
	return nil
}

// ReadLogs reads the logs into the context
func (ctx Context) ReadLogs(reader io.Reader) {
	defer trace.Tracef("")()

	ctx[logsKey] = readLogs(reader)
}

// AddLogs adds the logs into the context
func (ctx Context) AddLogs(logs []string) {
	defer trace.Tracef("")()

	log.Debugf("adding logs: %#v", logs)
	ctx[logsKey] = append(ctx.Logs(), logs...)
}

func readLogs(reader io.Reader) []string {
	defer trace.Tracef("")()

	scanner := bufio.NewScanner(reader)
	var logs []string
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	return logs
}
