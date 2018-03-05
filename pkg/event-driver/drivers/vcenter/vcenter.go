///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package vcenter

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/dispatch/pkg/event-driver"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// NO TESTS

const eventTypeVersion = "0.1"

type vCenterEvent struct {
	Metadata interface{} `json:"metadata"`
	Time     time.Time   `json:"time"`
	Category string      `json:"category"`
	Message  string      `json:"message"`
}

// NewConsumer creates a new vCenter event driver
func NewConsumer(vcenterURL string, insecure bool) (eventdriver.Consumer, error) {
	defer trace.Trace("")()
	vClient, err := newVCenterClient(context.Background(), vcenterURL, insecure)
	if err != nil {
		return nil, err
	}
	manager := event.NewManager(vClient.Client)
	return &vCenterDriver{
		vcenterURL: vcenterURL,
		insecure:   insecure,
		manager:    manager,
		client:     vClient,
	}, nil
}

type vCenterDriver struct {
	vcenterURL string
	insecure   bool
	manager    *event.Manager
	client     *govmomi.Client
	done       func()
}

func (d *vCenterDriver) Consume(topics []string) (<-chan *events.CloudEvent, error) {
	defer trace.Trace("")()
	ctx, cancel := context.WithCancel(context.Background())
	d.done = cancel
	eventsChan := make(chan *events.CloudEvent)
	go func() {
		defer trace.Trace("Consume loop")()
		err := d.manager.Events(
			ctx, // context
			// TODO: add support for filter customization
			[]types.ManagedObjectReference{d.client.ServiceContent.RootFolder}, // object(s) to monitor
			10,   // maximum number of events per page passed to handler
			true, // poll for events indefinitely
			true, // ignore limit of monitored objects (10)
			d.handler(eventsChan, false), // handler executed for each event page
		)
		if err != nil {
			log.Errorf("Error when reading events from vCenter: %+v", err)
		}
		close(eventsChan)
	}()

	return eventsChan, nil
}

func (d *vCenterDriver) Topics() []string {
	// TODO: generate it based on API WSDL
	return nil
}

func (d *vCenterDriver) Close() error {
	defer trace.Trace("")()
	d.done()
	return nil
}

func (d *vCenterDriver) handler(events chan *events.CloudEvent, multiple bool) func(types.ManagedObjectReference, []types.BaseEvent) error {
	defer trace.Trace("")()

	return func(obj types.ManagedObjectReference, page []types.BaseEvent) error {
		defer trace.Trace("actual handler")()

		event.Sort(page) // sort by event time

		for _, e := range page {
			processedEvent, err := d.processEvent(e)
			if err != nil {
				log.Errorf("error processing event: %+v", err)
			}

			events <- processedEvent
		}

		return nil
	}
}

func (d *vCenterDriver) processEvent(e types.BaseEvent) (*events.CloudEvent, error) {
	eventType := reflect.TypeOf(e).Elem().Name()
	defer trace.Tracef("event: %s", eventType)()

	log.Debugf("got event of type %s", eventType)

	cat, err := d.manager.EventCategory(context.Background(), e)
	if err != nil {
		log.Errorf("Error retrieving event category: %+v", err)
		return nil, err
	}

	ve := &vCenterEvent{
		Time:     e.GetEvent().CreatedTime,
		Category: cat,
		Message:  strings.TrimSpace(e.GetEvent().FullFormattedMessage),
	}

	// if this is a TaskEvent gather a little more information
	if t, ok := e.(*types.TaskEvent); ok {
		// some tasks won't have this information, so just use the event message
		if t.Info.Entity != nil {
			ve.Message = fmt.Sprintf("%s (target=%s %s)", ve.Message, t.Info.Entity.Type, t.Info.EntityName)
		}
	}
	ve.Metadata = processEventMetadata(e)

	topic := convertToTopic(eventType)

	return d.dispatchEvent(topic, ve)
}

func (d *vCenterDriver) dispatchEvent(topic string, ve *vCenterEvent) (*events.CloudEvent, error) {
	defer trace.Tracef("topic: %s", topic)()

	encoded, err := json.Marshal(*ve)
	if err != nil {
		return nil, err
	}

	event := events.CloudEvent{
		Namespace:          "vcenter.vmware.com",
		EventType:          topic,
		EventTypeVersion:   eventTypeVersion,
		CloudEventsVersion: events.CloudEventsVersion,
		SourceType:         "vcenter",
		SourceID:           "vcenter1", // TODO: make this unique
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Time{},
		ContentType:        "application/json",
		Data:               string(encoded),
	}

	return &event, nil
}

func newVCenterClient(ctx context.Context, vcenterURL string, insecure bool) (*govmomi.Client, error) {
	defer trace.Trace("")()

	url, err := soap.ParseURL(vcenterURL)
	if err != nil {
		return nil, err
	}

	return govmomi.NewClient(ctx, url, insecure)
}

func convertToTopic(eventType string) string {
	defer trace.Tracef("eventType: %s", eventType)()

	eventType = strings.Replace(eventType, "Event", "", -1)
	return camelCaseToDotSeparated(eventType)
}

func camelCaseToDotSeparated(src string) (topic string) {
	defer trace.Tracef("src: %s", src)()
	var words []string
	l := 0
	for s := src; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}
		words = append(words, strings.ToLower(s[:l]))
	}
	return strings.Join(words, ".")
}
