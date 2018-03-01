///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventdriver

import (
	"github.com/vmware/dispatch/pkg/events/driverclient"
	"github.com/vmware/dispatch/pkg/trace"
)

// NO TESTS

// New creates a new event driver
func New(client driverclient.Client, consumer Consumer) (Driver, error) {
	defer trace.Trace("")()
	return &defaultDriver{
		client:   client,
		consumer: consumer,
	}, nil
}

type defaultDriver struct {
	client   driverclient.Client
	consumer Consumer
}

func (driver *defaultDriver) Run() error {
	defer trace.Trace("")()
	eventsChan, err := driver.consumer.Consume(nil)
	if err != nil {
		return err
	}
	for event := range eventsChan {
		err = driver.client.SendOne(event)
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
	return nil
}
