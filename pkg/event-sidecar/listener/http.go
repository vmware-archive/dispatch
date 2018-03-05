///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package listener

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

// HTTPListener implements EventListener using HTTP server
type HTTPListener struct {
	SharedListener
	server *http.Server
	done   chan struct{}
}

// NewHTTP creates new HTTP Listener
func NewHTTP(shared SharedListener, port int) (*HTTPListener, error) {
	l := &HTTPListener{}

	l.server = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           l,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
	}
	l.SharedListener = shared
	l.done = make(chan struct{})

	return l, nil
}

// Serve starts serving loop. Blocks until Shutdown() is called.
func (l *HTTPListener) Serve() error {
	log.Printf("Listening on http://%s\n", l.server.Addr)
	err := l.server.ListenAndServe()
	if err == http.ErrServerClosed {
		// server is doing graceful shutdown, wait for it to finish
		<-l.done
		return nil
	}
	return err
}

// Shutdown gracefully shuts down the server.
func (l *HTTPListener) Shutdown() error {
	log.Printf("Shutting down...")
	err := l.server.Shutdown(context.Background())
	l.done <- struct{}{}
	return err
}

func (l *HTTPListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("Incorrect method %s, POST must be used", r.Method), http.StatusMethodNotAllowed)
		return
	}

	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		// This will be common as most drivers won't provide headers. No logging to reduce noise.
	}

	// Create the span referring to the parent.
	// If wireContext == nil, a root span will be created.
	serverSpan := opentracing.StartSpan("EventSidecar.ServeHTTP", opentracing.ChildOf(wireContext))
	defer serverSpan.Finish()
	spCtx := opentracing.ContextWithSpan(context.Background(), serverSpan)

	evs, err := l.parser.Parse(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing input: %s", err), http.StatusBadRequest)
		return
	}

	if len(evs) == 0 {
		http.Error(w, "No events parsed", http.StatusBadRequest)
		return
	}

	for _, ev := range evs {
		log.Debugf("Validating event %+v", ev)
		err = l.validator.Validate(&ev)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error validating event with ID %s: %s", ev.EventID, err), http.StatusBadRequest)
			return
		}
		log.Debugf("Pushing event %+v using topic %s and tenant %s", ev, ev.DefaultTopic(), l.tenant)
		err = l.transport.Publish(spCtx, &ev, ev.DefaultTopic(), l.tenant)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error publishing event with ID %s: %s", ev.EventID, err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
