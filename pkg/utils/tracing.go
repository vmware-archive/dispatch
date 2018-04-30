///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"io"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go/config"
)

type tracingOptions struct {
	opNameFunc func(r *http.Request) string
}

type nopCloser struct{}

// Close implements no-op io.Closer interface
func (nopCloser) Close() error {
	return nil
}

// TracingOption contols the behavior of the Middleware.
type TracingOption func(*tracingOptions)

// TODO: add more TracingOptions (for disabled, RPCMetrics, Tags, LogSpans, etc.)

// CreateTracer Returns a configured instance of opentracing.Tracer, a closer to close the tracer,
// and an error, if any.
func CreateTracer(serviceName string, hostPort string, opts ...TracingOption) (opentracing.Tracer, io.Closer, error) {
	if hostPort == "" {
		logrus.Warn("Tracer endpoint not provided, OpenTracing is disabled")
		return opentracing.NoopTracer{}, nopCloser{}, nil
	}
	cfg := config.Configuration{
		ServiceName: serviceName,
		Disabled:    false,
		RPCMetrics:  false,
		Tags:        nil,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  hostPort,
		},
	}
	return cfg.NewTracer(
		config.Logger(loggerAdapter{logrus.StandardLogger()}),
	)
}

// loggerAdapter wraps logrus to conform with jaeger.Logger interface
type loggerAdapter struct {
	logger logrus.FieldLogger
}

// Error logs an error message
func (l loggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

// Infof logs an info message
func (l loggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args)
}
