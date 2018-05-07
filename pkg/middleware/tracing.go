///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package middleware

import (
	"net/http"

	"github.com/justinas/alice"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// NO TESTS

// Tracer is a middleware that traces HTTP requests
type Tracer struct {
	tracer  opentracing.Tracer
	next    http.Handler
	options *mwOptions
}

type statusCodeTracker struct {
	http.ResponseWriter
	status int
}

// WriteHeader overwrites the default WriteHeader method of ResponseWriter
func (w *statusCodeTracker) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

type mwOptions struct {
	opNameFunc func(r *http.Request) string
}

// MWOption contols the behavior of the Middleware.
type MWOption func(*mwOptions)

// OperationNameFunc returns a MWOption that uses given function f
// to generate operation name for each server-side span.
func OperationNameFunc(f func(r *http.Request) string) MWOption {
	return func(options *mwOptions) {
		options.opNameFunc = f
	}
}

// NewTracingMW creates a new health check middleware at the specified path
func NewTracingMW(tr opentracing.Tracer, options ...MWOption) alice.Constructor {
	if tr == nil {
		// Disable tracing middleware when tracer is not available
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return NewTracer(tr, next, options...)
	}
}

// NewTracer creates a new health check middleware at the specified path
func NewTracer(tr opentracing.Tracer, next http.Handler, options ...MWOption) *Tracer {
	opts := mwOptions{
		opNameFunc: func(r *http.Request) string {
			return "HTTP " + r.Method + " " + r.URL.Path
		},
	}
	for _, opt := range options {
		opt(&opts)
	}

	return &Tracer{
		tracer:  tr,
		options: &opts,
		next:    next,
	}
}

// ServeHTTP is the middleware interface implementation
func (t *Tracer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx, _ := t.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	sp := t.tracer.StartSpan(t.options.opNameFunc(r), ext.RPCServerOption(ctx))
	ext.HTTPMethod.Set(sp, r.Method)
	ext.HTTPUrl.Set(sp, r.URL.String())
	ext.Component.Set(sp, "net/http")
	rw = &statusCodeTracker{rw, 200}
	r = r.WithContext(opentracing.ContextWithSpan(r.Context(), sp))

	t.next.ServeHTTP(rw, r)

	ext.HTTPStatusCode.Set(sp, uint16(rw.(*statusCodeTracker).status))
	sp.Finish()
}
