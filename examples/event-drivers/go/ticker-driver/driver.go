///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/satori/go.uuid"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/driverclient"
)

var seconds = flag.Int("seconds", 60, "Number of seconds to generate event after")
var debug = flag.Bool("debug", false, "Enable debug mode (print more information)")
var dryRun = flag.Bool("dryrun", false, "Enable dry run (does not send event")
var source = flag.String("source", uuid.NewV4().String(), "Set custom Source for the driver")
var endpoint = flag.String(driverclient.DispatchAPIEndpointFlag, "", "events api endpoint")

func main() {

	// Parse command line flags
	flag.Parse()

	// Get auth token
	token := os.Getenv(driverclient.AuthToken)

	// Use HTTP mode of sending events
	client, err := driverclient.NewHTTPClient(driverclient.WithEndpoint(*endpoint), driverclient.WithToken(token))
	if err != nil {
		panic(err)
	}

	// Create new ticker
	log.Printf("Creating ticker triggering every %d seconds", *seconds)
	duration := time.Duration(int64(*seconds)) * time.Second
	ticker := time.NewTicker(duration)

	// Catch SIGTERM signal to print nice shutdown message
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// Run infinite loop, waiting ether for ticker or for termination signal.
	// for every ticker tick, send an event (unless dry run mode is enabled).
	for {
		select {
		case <-ticker.C:
			go func() {
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
			}()
		case <-done:
			log.Printf("Shutting down...")
			return
		}
	}
}

func event() *events.CloudEvent {
	return &events.CloudEvent{
		EventType:          "ticker.tick",
		CloudEventsVersion: events.CloudEventsVersion,
		Source:             *source,
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Now(),
	}
}
