///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

// NO TESTS

import (
	"context"

	"github.com/opentracing-contrib/go-amqp/amqptracer"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// RabbitMQ implements transport over AMQP protocol and RabbitMQ messaging service
type RabbitMQ struct {
	url          string
	exchangeName string
	topicPrefix  string
	done         chan struct{}
	sendOnly     bool
	sendConn     *amqp.Connection
	recvConn     *amqp.Connection
}

// OptRabbitMQSendOnly creates only sending connection. Subscribe operation will panic
func OptRabbitMQSendOnly() func(mq *RabbitMQ) error {
	return func(mq *RabbitMQ) error {
		mq.sendOnly = true
		return nil
	}
}

// NewRabbitMQ creates new instance of RabbitMQ MessageQueue driver. Accepts
// variadic list of function options.
func NewRabbitMQ(url string, defaultExchangeName string, options ...func(mq *RabbitMQ) error) (mq *RabbitMQ, err error) {
	defer trace.Tracef("amqpurl: %+v, defaultExchangeName: %+v", url, defaultExchangeName)()

	mq = &RabbitMQ{
		url:          url,
		exchangeName: defaultExchangeName,
		done:         make(chan struct{}),
	}

	for _, option := range options {
		// TODO: handle errors from options
		option(mq)
	}
	// RabbitMQ docs:
	// "(...) it is advisable to only use individual AMQP connections for either producing or consuming."
	// https://www.rabbitmq.com/alarms.html
	mq.sendConn, err = amqp.Dial(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish a connection with RabbitMQ")
	}
	sendCloseChan := make(chan *amqp.Error)
	mq.sendConn.NotifyClose(sendCloseChan)
	go mq.shutdown(sendCloseChan)

	if !mq.sendOnly {
		mq.recvConn, err = amqp.Dial(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to establish a connection with RabbitMQ")
		}
		recvCloseChan := make(chan *amqp.Error)
		mq.recvConn.NotifyClose(recvCloseChan)
		go mq.shutdown(recvCloseChan)
	}

	ch, err := mq.sendConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to acquire a RabbitMQ channel")
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		defaultExchangeName, // name
		"topic",             // kind
		true,                // durable
		false,               // delete when unused
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to to declare an exchange")
	}

	return mq, nil
}

// Publish sends an event to RabbitMQ. tenant specifies the tenant
func (mq *RabbitMQ) Publish(ctx context.Context, event *events.CloudEvent, topic string, tenant string) error {
	defer trace.Tracef("topic: %s", event.EventType)()

	sp, _ := opentracing.StartSpanFromContext(
		ctx,
		"RabbitMQ.Publish",
	)
	defer sp.Finish()

	if mq.sendConn == nil {
		return errors.New("Connection not ready")
	}
	ch, err := mq.sendConn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to aquire a RabbitMQ channel")
	}
	defer ch.Close()

	if tenant == "" {
		tenant = mq.exchangeName
	}

	msg := mq.eventToMsg(event)
	// Inject the span context into the AMQP header.
	if err = amqptracer.Inject(sp, msg.Headers); err != nil {
		return err
	}

	err = ch.Publish(
		tenant,
		topic,
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return errors.Wrapf(err, "error when publishing a message, topic: %s, exchange: %s", event.EventType, tenant)
	}
	return nil
}

// Subscribe creates an active subscription on specified topic, and invokes handler function
// for every event received on given topic.
func (mq *RabbitMQ) Subscribe(ctx context.Context, topic string, handler events.Handler) (events.Subscription, error) {
	defer trace.Tracef("topic: %s", topic)()
	sp, _ := opentracing.StartSpanFromContext(
		ctx,
		"RabbitMQ.Subscribe",
	)
	defer sp.Finish()

	ch, q, err := mq.initQueue(topic)
	if err != nil {
		return nil, errors.Wrapf(err, "error initializing from queue %s", q.Name)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		true,   // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error when creating consume channel for queue %s", q.Name)
	}
	doneChan := make(chan struct{})
	go func() {
		defer trace.Tracef("listening for messages on topic: %s", topic)()
		for {
			select {
			case msg, open := <-msgs:
				if !open {
					ch.Close()
					return
				}
				spCtx, _ := amqptracer.Extract(msg.Headers)
				spSub := opentracing.StartSpan(
					"RabbitMQ.SubscriptionHandler",
					opentracing.FollowsFrom(spCtx),
				)

				// Update the context with the span for the subsequent reference.
				ctx = opentracing.ContextWithSpan(context.Background(), spSub)
				log.Debugf("Got an event: %s, %s, %s", msg.Exchange, msg.MessageId, msg.ContentType)
				event := mq.msgToEvent(msg)
				handler(ctx, event)
				msg.Ack(false)
				spSub.Finish()
			case <-doneChan:
				ch.Close()
				return
			}
		}
	}()
	return &subscription{done: doneChan}, nil
}

func (mq *RabbitMQ) eventToMsg(event *events.CloudEvent) amqp.Publishing {
	return amqp.Publishing{
		CorrelationId: event.SourceType,
		ContentType:   event.ContentType,
		ReplyTo:       event.SourceID,
		MessageId:     event.EventID,
		Timestamp:     event.EventTime,
		Type:          event.EventType,
		AppId:         event.Namespace,
		Body:          []byte(event.Data),
		Headers: amqp.Table{
			"dispatch-schema-url":         event.SchemaURL,
			"dispatch-event-type-version": event.EventTypeVersion,
		},
	}
}

func (mq *RabbitMQ) msgToEvent(message amqp.Delivery) *events.CloudEvent {
	return &events.CloudEvent{
		Namespace:          message.AppId,
		EventType:          message.Type,
		CloudEventsVersion: events.CloudEventsVersion,
		SourceType:         message.CorrelationId,
		SourceID:           message.ReplyTo,
		EventID:            message.MessageId,
		EventTime:          message.Timestamp,
		SchemaURL:          headerGet(message.Headers, "dispatch-schema-url"),
		ContentType:        message.ContentType,
		EventTypeVersion:   headerGet(message.Headers, "dispatch-event-type-version"),
		Data:               string(message.Body),
	}
}

// initQueue initializes and binds to a queue
func (mq *RabbitMQ) initQueue(topic string) (*amqp.Channel, *amqp.Queue, error) {
	ch, err := mq.recvConn.Channel()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to acquire a RabbitMQ channel")
	}
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error when declaring a queue")
	}

	err = ch.QueueBind(
		q.Name,          // queue name
		topic,           // routing key
		mq.exchangeName, // exchange
		false,           // noWait
		nil,             // args
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error when binding to a queue %s with topic %s and exchange %s", q.Name, topic, mq.exchangeName)
	}

	return ch, &q, nil
}

// Close closes AMQP connections and stops all subscriptions
func (mq *RabbitMQ) Close() {
	defer trace.Trace("")()
	if mq.sendConn != nil {
		mq.sendConn.Close()
	}
	if mq.recvConn != nil {
		mq.recvConn.Close()
	}
	if mq.done != nil {
		close(mq.done)
		mq.done = nil
	}
}

// shutdown is responsible for handling normal and abnormal rabbitMQ connection shutdown
func (mq *RabbitMQ) shutdown(c chan *amqp.Error) {
	defer trace.Trace("")()
	for {
		select {
		case err, ready := <-c:
			if !ready {
				// Graceful shutdown occurred
				log.Debug("RabbitMQ connection gracefully closed")
				return
			}
			// TODO: implement connection retry with exponential back-off
			log.Errorf("RabbitMQ connection error: %+v", err)
			return
		case <-mq.done:
			return
		}
	}
}

func headerGet(table amqp.Table, key string) string {
	if val, ok := table[key]; ok {
		switch val.(type) {
		case string:
			return val.(string)
		default:
			return ""
		}
	}
	return ""
}
