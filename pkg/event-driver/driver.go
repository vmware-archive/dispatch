///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventdriver

import (
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// NO TESTS

// New creates a new event driver
func New(queue events.Queue, consumer Consumer) (Driver, error) {
	defer trace.Trace("")()
	return &defaultDriver{
		queue:    queue,
		consumer: consumer,
	}, nil
}

type defaultDriver struct {
	queue    events.Queue
	consumer Consumer
}

func (driver *defaultDriver) Run() error {
	defer trace.Trace("")()
	eventsChan, err := driver.consumer.Consume(nil)
	if err != nil {
		return err
	}
	for event := range eventsChan {
		err = driver.queue.Publish(&event)
		if err != nil {
			// TODO: implement retry with exponential back-off
			return err
		}
	}
	return nil
}

func (driver *defaultDriver) Close() error {
	defer trace.Trace("")()
	driver.consumer.Close()
	driver.queue.Close()
	return nil
}
