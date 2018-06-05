///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package middleware

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/justinas/alice"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/version"
)

// NO TESTS

// HealthChecker is executed to verify health of the service.
type HealthChecker func() error

// HealthCheck is a middleware that serves healthcheck information
type HealthCheck struct {
	basePath string
	checker  HealthChecker
	next     http.Handler
}

// NewHealthCheckMW creates a new health check middleware at the specified path
func NewHealthCheckMW(basePath string, checker HealthChecker) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return NewHealthCheck(basePath, checker, next)
	}
}

// NewHealthCheck creates a new health check middleware at the specified path
func NewHealthCheck(basePath string, checker HealthChecker, next http.Handler) *HealthCheck {
	if basePath == "" {
		basePath = "/"
	}

	return &HealthCheck{
		basePath: basePath,
		checker:  checker,
		next:     next,
	}
}

type statusInfo struct {
	Version *v1.Version `json:"version"`
}

// ServeHTTP is the middleware interface implementation
func (h *HealthCheck) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, filepath.Join(h.basePath, "healthz")) {
		h.next.ServeHTTP(rw, r)
		return
	}

	var err error
	if h.checker != nil {
		err = h.checker()
	}

	var bs []byte
	if err == nil {
		bs, err = json.Marshal(&statusInfo{Version: version.Get()})
	}

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(bs)
}
