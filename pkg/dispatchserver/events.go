///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"io"
	"net/http"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/transport"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager"
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions"
)

type eventsConfig struct {
	Transport         string   `mapstructure:"transport" json:"transport,omitempty"`
	KafkaBrokers      []string `mapstructure:"kafka-brokers" json:"kafka-brokers,omitempty"`
	RabbitMQURL       string   `mapstructure:"rabbitmq-url" json:"rabbitmq-url,omitempty"`
	EventSidecarImage string   `mapstructure:"event-sidecar-image" json:"event-sidecar-image,omitempty"`
	K8sConfig         string   `mapstructure:"kubeconfig" json:"kubeconfig,omitempty"`
	K8sNamespace      string   `mapstructure:"namespace" json:"namespace,omitempty"`
	IngressHost       string   `mapstructure:"ingress-host" json:"ingress-host,omitempty"`
}

// NewCmdEvents creates a subcommand to run event manager
func NewCmdEvents(out io.Writer, config *serverConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "event-manager",
		Short:  i18n.T("Run Dispatch Event Manager"),
		Args:   cobra.NoArgs,
		PreRun: bindLocalFlags(&config.Events),
		Run: func(cmd *cobra.Command, args []string) {
			runEvents(config)
		},
	}
	cmd.SetOutput(out)

	cmd.Flags().String("transport", "kafka", "Event transport to use")
	cmd.Flags().StringSlice("kafka-brokers", []string{"localhost:9092"}, "host:port of Kafka broker(s)")
	cmd.Flags().String("rabbitmq-url", "amqp://guest:guest@localhost:5672/", "URL to RabbitMQ broker")
	cmd.Flags().String("event-sidecar-image", "", "Event sidecar image")
	cmd.Flags().String("kubeconfig", "", "Path to kubernetes config file")
	cmd.Flags().String("namespace", "default", "Kubernetes namespace")
	cmd.Flags().String("ingress-host", "", "Dispatch ingress hostname")
	return cmd
}

func runEvents(config *serverConfig) {
	store := entityStore(config)
	functions := functionsClient(config)
	secrets := secretsClient(config)

	var tr events.Transport
	var err error
	switch config.Events.Transport {
	case "kafka":
		tr, err = transport.NewKafka(config.Events.KafkaBrokers)
		if err != nil {
			log.Fatalf("Error creating Kafka event transport: %+v", err)
		}
	case "rabbitmq":
		tr, err = transport.NewRabbitMQ(config.Events.RabbitMQURL)
		if err != nil {
			log.Fatalf("Error creating RabbitMQ event transport: %+v", err)
		}
	default:
		log.Fatalf("Transport %s is not supported. pick one of [kafka,rabbitmq]", config.Events.Transport)
	}

	driverBackend, err := drivers.NewK8sBackend(
		secrets,
		drivers.ConfigOpts{
			SidecarImage:    config.Events.EventSidecarImage,
			TransportType:   config.Events.Transport,
			KafkaBrokers:    config.Events.KafkaBrokers,
			RabbitMQURL:     config.Events.RabbitMQURL,
			Tracer:          config.Tracer,
			K8sConfig:       config.Events.K8sConfig,
			DriverNamespace: config.Events.K8sNamespace,
			Host:            config.Events.IngressHost,
		},
	)
	if err != nil {
		log.Fatalf("Error creating k8sBackend: %v", err)
	}

	eventsDeps := eventsDependencies{
		store:           store,
		transport:       tr,
		driversBackend:  driverBackend,
		functionsClient: functions,
		secretsClient:   secrets,
	}

	eventsHandler, shutdown := initEvents(config, eventsDeps)
	defer shutdown()

	handler := addMiddleware(eventsHandler)
	server := httpServer(config)
	server.SetHandler(handler)
	defer server.Shutdown()
	if err := server.Serve(); err != nil {
		log.Error(err)
	}
}

type eventsDependencies struct {
	store           entitystore.EntityStore
	transport       events.Transport
	driversBackend  drivers.Backend
	functionsClient client.FunctionsClient
	secretsClient   client.SecretsClient
}

func initEvents(config *serverConfig, deps eventsDependencies) (http.Handler, func()) {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewEventManagerAPI(swaggerSpec)

	subManager, err := subscriptions.NewManager(deps.transport, deps.functionsClient)
	if err != nil {
		log.Fatalf("Error creating Event Subscription Manager: %v", err)
	}
	// event controller
	eventController := eventmanager.NewEventController(
		subManager,
		deps.driversBackend,
		deps.store,
		eventmanager.EventControllerConfig{
			ResyncPeriod: config.ResyncPeriod,
		},
	)

	eventController.Start()
	// handler
	handlers := &eventmanager.Handlers{
		Store:         deps.store,
		Transport:     deps.transport,
		Watcher:       eventController.Watcher(),
		SecretsClient: deps.secretsClient,
	}

	handlers.ConfigureHandlers(api)

	return api.Serve(nil), func() {
		eventController.Shutdown()
		deps.transport.Close()
	}
}
