///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package rabbitmq

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

type rabbitmq struct {
	url          string
	exchangeName string
	topicPrefix  string
	sendConn     *amqp.Connection
	recvConn     *amqp.Connection
	done         chan struct{}
}

// TopicPrefix sets prefix that will be applied to all events published by this driver
func TopicPrefix(prefix string) func(mq *rabbitmq) error {
	return func(mq *rabbitmq) error {
		mq.topicPrefix = prefix
		// TODO: validate length of prefix
		return nil
	}
}

// New creates new instance of RabbitMQ MessageQueue driver. Accepts
// variadic list of function options.
func New(url string, exchangeName string, options ...func(mq *rabbitmq) error) (events.Queue, error) {
	defer trace.Tracef("amqpurl: %+v, exchangeName: %+v", url, exchangeName)()
	sendConn, err := amqp.Dial(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish a connection with RabbitMQ")
	}
	recvConn, err := amqp.Dial(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish a connection with RabbitMQ")
	}
	mq := rabbitmq{
		url:          url,
		exchangeName: exchangeName,
		done:         make(chan struct{}),
	}

	for _, option := range options {
		// TODO: handle errors from options
		option(&mq)
	}

	mq.sendConn = sendConn
	mq.recvConn = recvConn

	sendCloseChan := make(chan *amqp.Error)
	recvCloseChan := make(chan *amqp.Error)
	mq.sendConn.NotifyClose(sendCloseChan)
	mq.recvConn.NotifyClose(recvCloseChan)
	go mq.shutdown(sendCloseChan)
	go mq.shutdown(recvCloseChan)

	ch, err := mq.sendConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to aquire a RabbitMQ channel")
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // kind
		false,        // durable
		false,        // delete when unused
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to to declare a queue")
	}

	return &mq, nil
}

// Publish sends an event to RabbitMQ on specified topic
func (mq *rabbitmq) Publish(ctx context.Context, event *events.Event) error {
	defer trace.Tracef("topic: %s", event.Topic)()

	sp := opentracing.SpanFromContext(ctx)
	defer sp.Finish()

	if mq.sendConn == nil {
		return errors.New("Connection not ready")
	}

	ch, err := mq.sendConn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to aquire a RabbitMQ channel")
	}
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: event.ContentType,
		Body:        event.Body,
		MessageId:   event.ID,
	}
	// Inject the span context into the AMQP header.
	if err := amqptracer.Inject(sp, msg.Headers); err != nil {
		return err
	}


	err = ch.Publish(
		mq.exchangeName,
		mq.topicPrefix+event.Topic,
		false, // mandatory
		false, // immediate
		,
	)
	if err != nil {
		return errors.Wrap(err, "error when publishing a message")
	}
	return nil
}

// Subscribe creates an active subscription on specified topic, and invokes handler function
// for every event received on given topic.
func (mq *rabbitmq) Subscribe(topic string, handler events.Handler) (events.Subscription, error) {
	defer trace.Tracef("topic: %s", topic)()
	ch, err := mq.recvConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to aquire a RabbitMQ channel")
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
		return nil, errors.Wrap(err, "error when declaring a queue")
	}

	err = ch.QueueBind(
		q.Name,          // queue name
		topic,           // routing key
		mq.exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error when binding to a queue %s with topic %s and exchange %s", q.Name, topic, mq.exchangeName)
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
		return nil, errors.Wrapf(err, "error when consuming messages from queue %s", q.Name)
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
				log.Debugf("Got an event: %s, %s, %s", msg.Exchange, msg.MessageId, msg.ContentType)
				event := events.Event{
					Topic:       topic,
					ID:          msg.MessageId,
					ContentType: msg.ContentType,
					Body:        msg.Body,
				}
				handler(&event)
				msg.Ack(false)
			case <-doneChan:
				ch.Close()
				return
			}
		}
	}()
	return &subscription{done: doneChan}, nil
}

func (mq *rabbitmq) Close() {
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
func (mq *rabbitmq) shutdown(c chan *amqp.Error) {
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
