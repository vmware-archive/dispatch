///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package trace

// NO TEST

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var tracingEnabled = true

// Enable global tracing.
func Enable() {
	tracingEnabled = true
}

// Disable global tracing.
func Disable() {
	tracingEnabled = false
}

// Logger instance used by trace to log messages.
var Logger = logrus.StandardLogger()

// Message is a trace object used to grab run-time state
type Message struct {
	msg      string
	funcName string
	lineNo   int

	startTime time.Time
}

func (t *Message) delta() time.Duration {
	if t == nil {
		return 0
	}
	return time.Now().Sub(t.startTime)
}

// begin a trace from this stack frame less the skip.
func newTrace(msg string, skip int) *Message {
	pc, _, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	name := runtime.FuncForPC(pc).Name()

	return &Message{
		msg:       msg,
		funcName:  name,
		lineNo:    line,
		startTime: time.Now(),
	}
}

// Trace encapsulates begin and end
// can be called like: defer trace.Trace("method name")()
func Trace(msg string) func() {
	if !tracingEnabled {
		return func() {}
	}
	tr := Begin(msg)
	return func() { End(tr) }
}

func Tracef(format string, a ...interface{}) func() {
	if !tracingEnabled {
		return func() {}
	}
	tr := Begin(fmt.Sprintf(format, a...))
	return func() { End(tr) }
}

// Begin starts the trace.  Msg is the msg to log.
func Begin(msg string) *Message {
	t := newTrace(msg, 3)
	if t == nil {
		return nil
	}
	if msg == "" {
		Logger.Debugf("[BEGIN] [%s:%d]", t.funcName, t.lineNo)
	} else {
		Logger.Debugf("[BEGIN] [%s:%d] %s", t.funcName, t.lineNo, t.msg)
	}
	return t
}

// End ends the trace.
func End(t *Message) {
	Logger.Debugf("[END  ] [%s:%d] [%s] %s", t.funcName, t.lineNo, t.delta(), t.msg)
}
