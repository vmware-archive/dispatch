///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"bufio"
	"io"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// Function context constants
const (
	LogsKey        = "logs"
	EventKey       = "event"
	ErrorKey       = "error"
	HTTPContextKey = "httpContext"
	// GoContext key is used for data that is not propagated to the FaaS function. it uses
	// context.Context underneath.
	GoContext = "goContext"
)

// Logs returns the logs as a list of strings
func (ctx Context) Logs() v1.Logs {
	log.Debugf(`Logs from ctx["logs"]: %#v`, ctx[LogsKey])
	switch logs := ctx[LogsKey].(type) {
	case v1.Logs:
		return logs
	case map[string]interface{}:
		var stdout []string
		var stderr []string

		if stdo, ok := logs["stdout"]; ok {
			if o, ok := stdo.([]interface{}); ok {
				for _, l := range o {
					s, ok := l.(string)
					if !ok {
						break
					}
					stdout = append(stdout, s)
				}
			}
		}

		if stde, ok := logs["stderr"]; ok {
			if e, ok := stde.([]interface{}); ok {
				for _, l := range e {
					s, ok := l.(string)
					if !ok {
						break
					}
					stderr = append(stderr, s)
				}
			}
		}

		return v1.Logs{Stderr: stderr, Stdout: stdout}
	}

	return v1.Logs{}
}

// ReadLogs reads the logs into the context
func (ctx Context) ReadLogs(stderrReader io.Reader, stdoutReader io.Reader) {
	ctx[LogsKey] = v1.Logs{
		Stderr: readLogs(stderrReader),
		Stdout: readLogs(stdoutReader),
	}
}

// AddLogs adds the logs into the context
func (ctx Context) AddLogs(logs v1.Logs) {
	log.Debugf("adding logs: %#v", logs)
	l := ctx.Logs()
	l.Stderr = append(l.Stderr, logs.Stderr...)
	l.Stdout = append(l.Stdout, logs.Stdout...)
	ctx[LogsKey] = l
}

func readLogs(reader io.Reader) []string {
	scanner := bufio.NewScanner(reader)
	var logs []string
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	return logs
}

// GetError returns the error from the context
func (ctx Context) GetError() *v1.InvocationError {
	switch invocationError := ctx[ErrorKey].(type) {
	case *v1.InvocationError:
		return invocationError
	case map[string]interface{}:
		var e v1.InvocationError
		if err := mapstructure.Decode(invocationError, &e); err == nil {
			return &e
		}
	}

	return nil
}

// SetError sets the error into the context
func (ctx Context) SetError(invocationError *v1.InvocationError) {

	ctx[ErrorKey] = invocationError
}
