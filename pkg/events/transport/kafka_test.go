///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/events"
)

func TestPublish(t *testing.T) {
	producer := mocks.NewSyncProducer(t, nil)
	event := events.NewCloudEventWithDefaults("dispatch")
	producer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(input []byte) error {
		var e events.CloudEvent
		err := json.Unmarshal(input, &e)
		if err != nil {
			return err
		}
		if e.EventID != event.EventID {
			return errors.New("not equal")
		}
		return nil
	})
	kafka := &Kafka{
		producer: producer,
	}

	err := kafka.Publish(context.Background(), &event, "dispatch", "dispatch")
	assert.NoError(t, err)
	err = producer.Close()
	assert.NoError(t, err)
}

func TestSubscribe(t *testing.T) {
	consumer := mocks.NewConsumer(t, nil)
	pc := consumer.ExpectConsumePartition("dispatch", 0, sarama.OffsetNewest)
	pc.ExpectMessagesDrainedOnClose()

	event := events.NewCloudEventWithDefaults("dispatch")
	eventBytes, _ := json.Marshal(event)
	pc.YieldMessage(&sarama.ConsumerMessage{
		Value: eventBytes,
	})

	kafka := &Kafka{
		consumer: consumer,
	}

	done := make(chan struct{})

	_, err := kafka.Subscribe(context.Background(), "dispatch", func(ctx context.Context, e *events.CloudEvent) {
		assert.Equal(t, event.EventID, e.EventID)
		done <- struct{}{}
	})
	assert.NoError(t, err)

	<-done

}
