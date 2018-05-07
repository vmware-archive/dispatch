///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

import (
	"context"
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// Kafka Implements transport interface using Kafka broker.
type Kafka struct {
	producer     sarama.SyncProducer
	consumer     sarama.Consumer
	producerOnly bool
}

// OptKafkaSendOnly creates producer only. Subscribe operation will panic
func OptKafkaSendOnly() func(k *Kafka) error {
	return func(k *Kafka) error {
		k.producerOnly = true
		return nil
	}
}

// NewKafka creates an instance of transport based on Kafka broker.
func NewKafka(brokerAddrs []string, options ...func(k *Kafka) error) (*Kafka, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true
	syncProducer, err := sarama.NewSyncProducer(brokerAddrs, config)
	if err != nil {
		return nil, err
	}

	k := Kafka{
		producer: syncProducer,
	}
	for _, option := range options {
		// TODO: handle errors from options
		option(&k)
	}
	if k.producerOnly {
		return &k, nil
	}
	config = sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	k.consumer, err = sarama.NewConsumer(brokerAddrs, config)
	return &k, err
}

// Publish publishes an event
func (k *Kafka) Publish(ctx context.Context, event *events.CloudEvent, topic string, tenant string) error {
	defer trace.Tracef("topic: %s", event.EventType)()
	sp, _ := opentracing.StartSpanFromContext(
		ctx,
		"Kafka.Publish",
	)
	defer sp.Finish()

	msg, err := fromEvent(event)
	if err != nil {
		return errors.Wrapf(err, "error when creating Kafka message from CloudEvent for topic %s", topic)
	}
	msg.Topic = topic

	err = injectSpan(sp, msg)
	if err != nil {
		return errors.Wrap(err, "error injecting opentracing span to Kafka message")
	}
	if _, _, err = k.producer.SendMessage(msg); err != nil {
		return errors.Wrap(err, "error sending Kafka message")
	}

	return nil
}

// Subscribe subscribes to an event
func (k *Kafka) Subscribe(ctx context.Context, topic string, handler events.Handler) (events.Subscription, error) {
	defer trace.Tracef("topic: %s", topic)()
	sp, _ := opentracing.StartSpanFromContext(
		ctx,
		"Kafka.Subscribe",
	)
	defer sp.Finish()

	doneChan := make(chan struct{})

	// create partition consumer. Since we are creating only one consumer, we should automatically consume messages
	// from all partitions regardless of the partition number we select
	partitionConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating a partition consumer for topic %s", topic)
	}

	go func() {
		defer trace.Tracef("listening for messages on topic: %s", topic)()
		for {
			select {
			case msg, open := <-partitionConsumer.Messages():
				if !open {
					partitionConsumer.Close()
					return
				}
				log.Debugf("Received a kafka message %+v", *msg)
				spCtx, err := extractSpan(msg)
				if err != nil {
					log.Debugf("Unable to extract tracing span for message on topic %s: %+v", msg.Topic, err)
				}
				spSub := opentracing.StartSpan(
					"Kafka.SubscriptionHandler",
					opentracing.FollowsFrom(spCtx),
				)
				func() {
					defer spSub.Finish()
					// Update the context with the span for the subsequent reference.
					ctx = opentracing.ContextWithSpan(context.Background(), spSub)
					event, err := toEvent(msg)
					if err != nil {
						log.Errorf("Error when converting Kafka message to event: %+v", err)
						return
					}
					log.Debugf("Got an event %+v", event)
					handler(ctx, event)
				}()

			case <-doneChan:
				partitionConsumer.Close()
				return
			}
		}
	}()

	return &subscription{done: doneChan}, nil
}

// Close closes the transport
func (k *Kafka) Close() {
	if err := k.producer.Close(); err != nil {
		log.Warnf("error when closing Kafka producer: %+v", err)
	}
	if k.consumer == nil {
		return
	}
	if err := k.consumer.Close(); err != nil {
		log.Warnf("error when closing Kafka consumer: %+v", err)
	}
}

// injectSpan injects OpenTracing Span into sarama.ProducerMessage.Headers structure.
func injectSpan(span opentracing.Span, message *sarama.ProducerMessage) error {
	headers := kafkaProducerMsgHeaders(message.Headers)
	if err := span.Tracer().Inject(span.Context(), opentracing.TextMap, &headers); err != nil {
		return err
	}
	message.Headers = headers
	return nil
}

// extractSpan extracts OpenTracing Span from sarama.ConsumerMessage.Headers structure.
func extractSpan(message *sarama.ConsumerMessage) (opentracing.SpanContext, error) {
	headers := kafkaConsumerMsgHeaders(message.Headers)
	return opentracing.GlobalTracer().Extract(opentracing.TextMap, &headers)
}

// kafkaProducerMsgHeaders implements opentracing.TextMapWriter interface
type kafkaProducerMsgHeaders []sarama.RecordHeader

// Set sets value val for given key.
func (h *kafkaProducerMsgHeaders) Set(key, val string) {
	// This has an obvious problem of O(n) complexity in pessimistic scenario.
	// OpenTracing usually sets only few headers, if it becomes a performance issue we can build
	// a map or just append without checking.
	for i, rec := range *h {
		if string(rec.Key) == key {
			(*h)[i].Value = []byte(val)
			return
		}
	}
	*h = append(*h, sarama.RecordHeader{Key: []byte(key), Value: []byte(val)})
}

// kafkaConsumerMsgHeaders implements opentracing.TextMapReader interface
type kafkaConsumerMsgHeaders []*sarama.RecordHeader

// ForeachKey executes handler for each key/value  tuple in headers
func (h *kafkaConsumerMsgHeaders) ForeachKey(handler func(key, val string) error) error {
	for _, rec := range *h {
		if err := handler(string(rec.Key), string(rec.Value)); err != nil {
			return err
		}
	}
	return nil
}

func toEvent(message *sarama.ConsumerMessage) (*events.CloudEvent, error) {
	event := &events.CloudEvent{}
	err := json.Unmarshal(message.Value, event)
	return event, err
}

func fromEvent(event *events.CloudEvent) (*sarama.ProducerMessage, error) {
	data, err := json.Marshal(event)
	return &sarama.ProducerMessage{Value: sarama.ByteEncoder(data)}, err
}
