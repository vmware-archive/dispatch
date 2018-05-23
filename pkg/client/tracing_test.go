///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/client"
)

func TestTracingRoundTrip(t *testing.T) {
	fakeTracer := mocktracer.New()
	fakeSpan := fakeTracer.StartSpan("fakeOperation")
	opentracing.SetGlobalTracer(fakeTracer)
	rt := client.NewTracingRoundTripper(nil)
	r, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Error(err)
	}
	r = r.WithContext(opentracing.ContextWithSpan(context.Background(), fakeSpan))
	_, err = rt.RoundTrip(r)
	assert.Error(t, err)
	assert.True(t, len(r.Header) > 0)

	extractedSpan, err := fakeTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, extractedSpan)
}
