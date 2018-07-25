///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package riff

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/utils"
)

type returns struct {
	sync.Mutex
	m map[string]chan message.Message
}

func newReturns() *returns {
	return &returns{m: make(map[string]chan message.Message)}
}

func (r *returns) Put(k string, c chan message.Message) {
	r.Lock()
	defer r.Unlock()
	r.m[k] = c
}

func (r *returns) Remove(k string) chan message.Message {
	r.Lock()
	defer r.Unlock()
	defer delete(r.m, k)
	return r.m[k]
}

type requester struct {
	requestIDKey string

	timeout time.Duration

	returns *returns

	producer transport.Producer
	consumer transport.Consumer

	done chan struct{}
}

func newRequester(requestIDKey string, producer transport.Producer, consumer transport.Consumer) *requester {
	r := &requester{
		requestIDKey: requestIDKey,
		timeout:      defaultTimeout,
		returns:      newReturns(),
		producer:     producer,
		consumer:     consumer,
		done:         make(chan struct{}),
	}
	go r.run()
	return r
}

func (r *requester) Close() error {
	close(r.done)
	return nil
}

func (r *requester) run() {
	defer utils.Close(r.consumer)
	defer utils.Close(r.producer)

	for {
		select {
		case msg := <-r.consumer.Messages():
			s := msg.Headers()[r.requestIDKey]
			if len(s) != 1 {
				log.Warnf("could not get requestID from message: %+v", msg)
				continue
			}
			requestID := s[0]
			resultChan := r.returns.Remove(requestID)
			if resultChan == nil {
				log.Errorln("Most likely that the function was created in a different pod.")
				log.Errorf("cannot find resultChan for requestID: '%s', msg: %+v", requestID, msg)
				continue
			}

			select {
			case resultChan <- msg:
			default:
				log.Errorf("error sending to resultChan '%v', requestID: '%s', msg: %+v", resultChan, requestID, msg)
			}
		case <-r.done:
			return
		}
	}
}

func (r *requester) logProducerErrors() {
	for err := range r.producer.Errors() {
		log.Errorf("%+v", err)
	}
}

// ContentType is a constant for Content-Type
const ContentType = "Content-Type"

// Accept is a constant for Accept
const Accept = "Accept"
const jsonContentType = "application/json"
const defaultTimeout = 5 * time.Minute

func (r requester) makeHeaders(runID string) message.Headers {
	return message.Headers{
		ContentType:    []string{jsonContentType},
		Accept:         []string{jsonContentType},
		r.requestIDKey: []string{runID},
	}
}

func (r *requester) Request(topic string, reqID string, payload []byte) ([]byte, error) {
	resultChan := make(chan message.Message)
	r.returns.Put(reqID, resultChan)

	if err := r.producer.Send(topic, message.NewMessage(payload, r.makeHeaders(reqID))); err != nil {
		return nil, errors.Wrapf(err, "riff driver: error sending to producer, reqID: %s", reqID)
	}

	timer := time.NewTimer(r.timeout)
	select {
	case msg := <-resultChan:
		return msg.Payload(), nil
	case <-timer.C:
		r.returns.Remove(reqID)
		return nil, errors.Errorf("timeout getting response from function, reqID: %s", reqID)
	}
}
