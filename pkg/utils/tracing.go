///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
)

// AddHTTPTracing returns an opentracing span with HTTP tracer information
func AddHTTPTracing(r *http.Request, operationName string) (opentracing.Span, context.Context) {
	wireContext, _ := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	// Create the span referring to the parent.
	// If wireContext == nil, a root span will be created.
	serverSpan := opentracing.StartSpan("EventManager.emitEvent", opentracing.ChildOf(wireContext))
	spCtx := opentracing.ContextWithSpan(r.Context(), serverSpan)
	return serverSpan, spCtx
}
