///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package trace

import (
	"context"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestTrace(t *testing.T) {
	mockTracer := &mocktracer.MockTracer{}
	opentracing.SetGlobalTracer(mockTracer)
	span, _ := Trace(context.Background(), "")
	span.Finish()
	spans := mockTracer.FinishedSpans()
	assert.Len(t, spans, 1)
	finishedSpan := spans[0]
	assert.Equal(t, "github.com/vmware/dispatch/pkg/trace.TestTrace", finishedSpan.OperationName)
	logs := finishedSpan.Logs()
	assert.Len(t, logs, 1)
	fields := logs[0].Fields
	assert.Len(t, fields, 2)
	assert.Equal(t, "calledFrom", fields[0].Key)
	assert.Equal(t, "caller", fields[1].Key)
	assert.Equal(t, "github.com/vmware/dispatch/pkg/trace.TestTrace", fields[1].ValueString)
}

func TestTraceCustomOpName(t *testing.T) {
	mockTracer := &mocktracer.MockTracer{}
	opentracing.SetGlobalTracer(mockTracer)
	span, _ := Trace(context.Background(), "testOperation")
	span.Finish()
	spans := mockTracer.FinishedSpans()
	assert.Len(t, spans, 1)
	finishedSpan := spans[0]
	assert.Equal(t, "testOperation", finishedSpan.OperationName)
	logs := finishedSpan.Logs()
	assert.Len(t, logs, 1)
	fields := logs[0].Fields
	assert.Len(t, fields, 2)
	assert.Equal(t, "calledFrom", fields[0].Key)
	assert.Equal(t, "caller", fields[1].Key)
	assert.Equal(t, "github.com/vmware/dispatch/pkg/trace.TestTraceCustomOpName", fields[1].ValueString)
}
