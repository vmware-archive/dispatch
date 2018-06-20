///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

import (
	"context"
	"fmt"
	"io"

	"github.com/vmware/dispatch/pkg/events"
)

// NO TESTS

// Noop implements dummy transport which does nothing (optionally prints to the output specified)
type Noop struct {
	out io.Writer
}

// NewNoop creates an instance of transport. out can be set to configure optional output.
func NewNoop(out io.Writer) *Noop {
	return &Noop{
		out: out,
	}
}

// Publish publishes event.
func (t *Noop) Publish(ctx context.Context, event *events.CloudEvent, topic string, organization string) error {
	if t.out != nil {
		fmt.Fprintf(t.out, "Event %+v published to topic %s and organization %s\n", event, topic, organization)
	}
	return nil
}

// Subscribe subscribes to event.
func (t *Noop) Subscribe(ctx context.Context, topic string, organization string, handler events.Handler) (events.Subscription, error) {
	if t.out != nil {
		fmt.Fprintf(t.out, "Subscription to topic %s using handler %T\n", topic, handler)
	}
	return nil, nil
}

// Close closes transport.
func (t *Noop) Close() {
	if t.out != nil {
		fmt.Fprintf(t.out, "Transport closed")
	}
}
