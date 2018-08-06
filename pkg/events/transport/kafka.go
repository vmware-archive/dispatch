///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// Kafka Implements transport interface using Kafka broker.
type Kafka struct {
	producer     sarama.SyncProducer
	consumers    []*cluster.Consumer
	config       *cluster.Config
	addrs        []string
	client       *cluster.Client
	partitions   int32
	producerOnly bool
}

// OptKafkaSendOnly creates producer only. Subscribe operation will panic
func OptKafkaSendOnly() func(k *Kafka) error {
	return func(k *Kafka) error {
		k.producerOnly = true
		return nil
	}
}

type offsetPartitioner struct {
	topic  string
	client *cluster.Client
}

// Partition allows offsetPartitioner to implement the cluster.Strategy interface, so that we control which partition
// a message will go to. In this case it will be sent to the partition that has the lowest offset. A partitions offset
// is global, so this provides some level of synchronization between event managers
func (p *offsetPartitioner) Partition(message *sarama.ProducerMessage, numPartitions int32) (int32, error) {
	min := math.Inf(1)
	partition := int32(-1)
	for i := int32(0); i < numPartitions; i++ {
		offset, err := p.client.GetOffset(p.topic, i, sarama.OffsetNewest)
		if err != nil {
			return -1, errors.Wrapf(err, "Unable to get offset for topic %s partition %v", p.topic, i)
		}
		if min >= float64(offset) {
			min = float64(offset)
			partition = i
		}
	}
	return partition, nil
}

// RequiresConsistency is also part of the cluster.Strategy interface. We want consistency between offsets.
func (p *offsetPartitioner) RequiresConsistency() bool {
	return true
}

// NewOffsetPartitioner returns a partitioner that partitions based on which partition has the lowest offset
// B/c all messages are ~ the same size => this will be least busy partition
func NewOffsetPartitioner(client *cluster.Client) func(topic string) sarama.Partitioner {
	return func(topic string) sarama.Partitioner {
		return &offsetPartitioner{
			client: client,
			topic:  topic,
		}
	}
}

// NewKafka creates an instance of transport based on Kafka broker.
func NewKafka(brokerAddrs []string, numClients int, options ...func(k *Kafka) error) (*Kafka, error) {
	config := cluster.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_11_0_0
	config.Group.PartitionStrategy = cluster.StrategyRoundRobin
	client, err := cluster.NewClient(brokerAddrs, config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to open connection to client")
	}
	log.Infof("Trying to create %v kafka clients", numClients)

	config.Producer.Partitioner = NewOffsetPartitioner(client)

	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	k := Kafka{
		producer:   syncProducer,
		addrs:      brokerAddrs,
		config:     config,
		partitions: int32(numClients),
		client:     client,
		consumers:  []*cluster.Consumer{},
	}
	for _, option := range options {
		// TODO: handle errors from options
		option(&k)
	}
	return &k, nil
}

// Publish publishes an event
func (k *Kafka) Publish(ctx context.Context, event *events.CloudEvent, topic string, organization string) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	log.Debugf("Received Cloud Event for topic %v", topic)

	if organization == "" {
		return errors.New("organization cannot be empty")
	}

	if topic == "" {
		return errors.New("topic cannot be empty")
	}

	topicWithOrg := organization + "." + topic

	msg, err := fromEvent(event)
	if err != nil {
		return errors.Wrapf(err, "error turning Kafka message into CloudEvent for topic %s in organization %s", topic, organization)
	}
	msg.Topic = topicWithOrg

	err = injectSpan(span, msg)
	if err != nil {
		return errors.Wrap(err, "error injecting opentracing span to Kafka message")
	}
	partition, _, err := k.producer.SendMessage(msg)
	if err != nil {
		return errors.Wrap(err, "error sending Kafka message")
	}
	log.Debugf("Sent a message on partition: %v", partition)

	return nil
}

func requestTopic(topic string, partitions int32) *sarama.CreateTopicsRequest {
	details := sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: 3,
	}
	return &sarama.CreateTopicsRequest{
		Version: 1,
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: &details,
		},
		Timeout:      5 * time.Second,
		ValidateOnly: false,
	}
}

// Subscribe subscribes to an event
func (k *Kafka) Subscribe(ctx context.Context, topic string, organization string, handler events.Handler) (events.Subscription, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	if organization == "" {
		return nil, errors.New("organization cannot be empty")
	}

	if topic == "" {
		return nil, errors.New("topic cannot be empty")
	}

	topicWithOrg := organization + "." + topic

	doneChan := make(chan struct{})

	controller, err := k.client.Controller()
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't get controller")
	}
	request := requestTopic(topicWithOrg, k.partitions)

	resp, err := controller.CreateTopics(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create topic: %v", topicWithOrg)
	}
	for topic, terr := range resp.TopicErrors {
		if terr.Err != sarama.ErrTopicAlreadyExists {
			log.Warnf("Topic %s has err %v", topic, *terr)
		} else {
			log.Debugf("Topic %s already exists!", topic)
		}
	}

	time.Sleep(5 * time.Second)

	log.Debugf("Created topic: %v", topicWithOrg)

	k.client.RefreshMetadata(topicWithOrg)

	consumer, err := cluster.NewConsumer(k.addrs, "dispatch-event-manager", []string{topicWithOrg}, k.config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create consumer")
	}

	k.consumers = append(k.consumers, consumer)

	// Consume Messages
	go func() {
		for {
			select {
			case msg, open := <-consumer.Messages():
				if !open {
					consumer.Close()
					return
				}
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
					handler(ctx, event)
					// Mark message as processed
					consumer.MarkOffset(msg, "")
				}()

			case <-doneChan:
				consumer.Close()
				return
			}
		}
	}()

	return &subscription{done: doneChan, topic: topic, organization: organization}, nil
}

// Close closes the transport
func (k *Kafka) Close() {
	if err := k.producer.Close(); err != nil {
		log.Warnf("error when closing Kafka producer: %+v", err)
	}
	k.client.Close()
	for _, consumer := range k.consumers {
		err := consumer.Close()
		if err != nil {
			log.Warnf("error when closing cluster client: %+v", err)
		}
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
