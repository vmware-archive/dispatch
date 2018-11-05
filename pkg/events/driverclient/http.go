///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package driverclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/validator"
	"github.com/vmware/dispatch/pkg/utils"
)

const (
	// DispatchAPIEndpointFlag defines the name of the flag used by Event Manager to set its API endpoint
	// when provisioning driver. Drivers must implement this flag and use WithEndpoint with its value.
	DispatchAPIEndpointFlag = "dispatch-api-endpoint"

	// AuthToken defines the environment variable set by Event manager to authenticate events.
	// Drivers must read this environment variable and use WithToken if value is is set.
	AuthToken = "AUTH_TOKEN"

	eventsAPIPath  = "v1/event/ingest"
	authTokenKey   = "authToken"
	defaultAPIHost = "localhost"
	defaultAPIPort = 8080
)

// HTTPClientOpt allows customization of HTTPClient
type HTTPClientOpt func(client *HTTPClient) error

// WithEndpoint allows to customize host & port
func WithEndpoint(endpoint string) HTTPClientOpt {
	return func(client *HTTPClient) error {
		if endpoint == "" {
			return errors.New("missing value for endpoint")
		}
		hostPort := strings.Split(endpoint, ":")
		client.host = hostPort[0]
		if len(hostPort) > 1 {
			port, err := strconv.Atoi(hostPort[1])
			if err != nil {
				return err
			}
			client.port = port
		}
		return nil
	}
}

// WithPort allows to customize port
func WithPort(port string) HTTPClientOpt {
	return func(client *HTTPClient) error {
		if port == "" {
			port = "8080"
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		client.port = p
		return nil
	}
}

// WithHost allows to customize host
func WithHost(host string) HTTPClientOpt {
	return func(client *HTTPClient) error {
		if host == "" {
			ips, err := net.LookupIP("host.docker.internal")
			if err != nil {
				return err
			}
			host = ips[0].String()
		}
		client.host = host
		return nil
	}
}

// WithTracer allows setting custom tracer
func WithTracer(t opentracing.Tracer) HTTPClientOpt {
	return func(client *HTTPClient) error {
		client.tracer = t
		return nil
	}
}

// WithToken allows setting custom authentication token
func WithToken(token string) HTTPClientOpt {
	return func(client *HTTPClient) error {
		client.authToken = token
		return nil
	}
}

// HTTPClient implements event driver client using HTTP protocol
type HTTPClient struct {
	client    *http.Client
	host      string
	port      int
	endpoint  string
	validator events.Validator
	authToken string

	tracer opentracing.Tracer
}

// NewHTTPClient returns new instance of driverclient.Client using HTTPClient implementation
func NewHTTPClient(opts ...HTTPClientOpt) (Client, error) {
	c := &HTTPClient{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		host:      defaultAPIHost,
		port:      defaultAPIPort,
		endpoint:  eventsAPIPath,
		tracer:    opentracing.NoopTracer{},
		validator: validator.NewDefaultValidator(),
	}

	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	return c, c.checkHealth()
}

// Send sends slice of vents to Dispatch system. It runs Validate() first.
func (c *HTTPClient) Send(evs []events.CloudEvent) error {
	if err := c.Validate(evs); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(evs)
	if err != nil {
		return err
	}

	return c.send(buf)
}

// SendOne sends single event to Dispatch system. It runs Validate() first.
func (c *HTTPClient) SendOne(event *events.CloudEvent) error {
	if err := c.ValidateOne(event); err != nil {
		return err
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(eventJSON)

	return c.send(buf)
}

// Validate validates slice of events without sending it
func (c *HTTPClient) Validate(evs []events.CloudEvent) error {
	for _, e := range evs {
		if err := c.validator.Validate(&e); err != nil {
			return err
		}
	}
	return nil
}

// ValidateOne validates single event without sending it
func (c *HTTPClient) ValidateOne(event *events.CloudEvent) error {
	return c.validator.Validate(event)
}

func (c *HTTPClient) getURL() string {
	return fmt.Sprintf("http://%s:%d/%s", c.host, c.port, c.endpoint)
}

func (c *HTTPClient) send(buf *bytes.Buffer) error {
	req, _ := http.NewRequest("POST", c.getURL(), buf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", "unset")

	if c.authToken != "" {
		q := req.URL.Query()
		q.Add(authTokenKey, c.authToken)
		req.URL.RawQuery = q.Encode()
	}

	// TODO(karols): add debug flag handling
	_, err := c.client.Do(req)
	return err
}

func (c *HTTPClient) checkHealth() error {
	return utils.Backoff(30*time.Second, func() error {
		log.Printf("checking connection to %s", c.getURL())
		_, err := c.client.Get(c.getURL())
		if err != nil {
			log.Printf("connection failed: %s", err)
		}
		return err
	})
}
