///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package trace

// NO TEST

import (
	"context"
	"fmt"
	"runtime"
	"strconv"

	"github.com/opentracing/opentracing-go"
)

// begin a span from this stack frame less the skip.
func newSpan(ctx context.Context, opName string, skip int) (opentracing.Span, context.Context) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil, context.Background()
	}
	caller := runtime.FuncForPC(pc).Name()
	if opName == "" {
		opName = caller
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, opName)
	span.LogKV(
		"calledFrom", file+":"+strconv.Itoa(line),
		"caller", caller,
	)

	return span, ctx
}

// Trace encapsulates begin and end
// can be called like: defer trace.Trace(context.Background(), "operation name").Finish()
func Trace(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	return newSpan(ctx, operationName, 2)
}

// Tracef encapsulates begin and end
// can be called like: defer trace.Tracef(context.Background(), "operation name %s", param).Finish()
// Like Trace but takes a format string and parameters.
func Tracef(ctx context.Context, format string, a ...interface{}) (opentracing.Span, context.Context) {
	return Trace(ctx, fmt.Sprintf(format, a...))
}
