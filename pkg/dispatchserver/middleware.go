///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"log"
	"net/http"

	"github.com/justinas/alice"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/vmware/dispatch/pkg/middleware"
	"github.com/vmware/dispatch/pkg/utils"
)

func addMiddleware(handler http.Handler) http.Handler {
	healthChecker := func() error {
		// TODO: implement service-specific healthchecking
		return nil
	}

	tracer, tracingCloser, err := utils.CreateTracer("EventManager", "")
	if err != nil {
		log.Fatalf("Error creating a tracer: %+v", err)
	}
	defer tracingCloser.Close()
	opentracing.SetGlobalTracer(tracer)

	return alice.New(
		middleware.NewHealthCheckMW("", healthChecker),
		middleware.NewTracingMW(tracer),
	).Then(handler)
}
