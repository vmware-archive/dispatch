///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

import (
	"context"
	"sync"

	"github.com/vmware/dispatch/pkg/events"
)

const (
	defaultQueueSize = 20
)

type queue map[string]chan events.CloudEvent

// InMemory provides event transport implemented completely in memory.
type InMemory struct {
	exchanges map[string]queue
	mu        sync.Mutex

	queueSize int
}

// NewInMemory returns an initialized instance of InMemory event transport.
func NewInMemory() *InMemory {
	return &InMemory{
		exchanges: make(map[string]queue),
		queueSize: defaultQueueSize,
	}
}

// Publish implements Transport interface publish method
func (m *InMemory) Publish(ctx context.Context, event *events.CloudEvent, topic string, organization string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	exchange, ok := m.exchanges[organization]
	if !ok || exchange == nil {
		return nil
	}
	queue, ok := exchange[topic]
	if !ok || queue == nil {
		return nil
	}
	queue <- *event

	return nil
}

// Subscribe implements Transport interface subscribe method
func (m *InMemory) Subscribe(ctx context.Context, topic string, organization string, handler events.Handler) (events.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exchange, ok := m.exchanges[organization]
	if !ok || exchange == nil {
		exchange = make(queue)
		m.exchanges[organization] = exchange
	}
	queue, ok := exchange[topic]
	if !ok || queue == nil {
		queue = make(chan events.CloudEvent, m.queueSize)
		exchange[topic] = queue
	}

	doneChan := make(chan struct{})
	go func() {
		for {
			select {
			case event := <-queue:
				handler(ctx, &event)
			case <-doneChan:
				m.mu.Lock()
				close(queue)
				delete(exchange, topic)
				m.mu.Unlock()
				return
			}
		}
	}()
	return &subscription{done: doneChan, topic: topic, organization: organization}, nil
}

// Close implements Transport interface close method.
func (m *InMemory) Close() {

}
