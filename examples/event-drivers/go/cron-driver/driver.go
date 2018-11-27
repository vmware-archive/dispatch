///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron"

	"github.com/satori/go.uuid"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/driverclient"
)

var spec = flag.String("cron", "59 59 23 31 12 ? 2099", "cron expression for sending cloudevents on a schedule")
var debug = flag.Bool("debug", false, "Enable debug mode (print more information)")
var dryRun = flag.Bool("dryrun", false, "Enable dry run (does not send event")
var source = flag.String("source", uuid.NewV4().String(), "Set custom Source for the driver")
var gateway = flag.String(driverclient.DispatchEventsGatewayFlag, "", "events gateway")

func main() {

	// Parse command line flags
	flag.Parse()

	// Get auth token
	token := os.Getenv(driverclient.AuthToken)

	// Use HTTP mode of sending events
	client, err := driverclient.NewHTTPClient(driverclient.WithGateway(*gateway), driverclient.WithToken(token))
	if err != nil {
		panic(err)
	}

	sendFunc := func() {
		ev := event()
		if *debug {
			log.Printf("Sending event %+v", *ev)
		}
		if !*dryRun {
			err := client.SendOne(ev)
			if err != nil {
				log.Printf("Error sending event: %s", err)
			}
		}
	}

	log.Printf("Creating new scheduled event with cron spec %s", *spec)
	c := cron.New()
	err = c.AddFunc(*spec, sendFunc)
	if err != nil {
		panic(err)
	}
	c.Start()

	// Catch SIGTERM signal to print nice shutdown message
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// Run infinite loop, waiting ether for ticker or for termination signal.
	// for every ticker tick, send an event (unless dry run mode is enabled).
	for {
		select {
		case <-done:
			log.Printf("Shutting down...")
			c.Stop()
			return
		}
	}
}

func event() *events.CloudEvent {
	return &events.CloudEvent{
		EventType:          "cron.trigger",
		CloudEventsVersion: events.CloudEventsVersion,
		Source:             *source,
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Now(),
	}
}
