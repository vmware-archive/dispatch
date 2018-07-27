///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventsidecar

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/vmware/dispatch/pkg/event-sidecar/listener"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/parser"
	"github.com/vmware/dispatch/pkg/events/transport"
	"github.com/vmware/dispatch/pkg/events/validator"
	"github.com/vmware/dispatch/pkg/utils"
)

// NO TESTS

type sidecarConfig struct {
	ListenerHTTPPort int      `mapstructure:"listener-http-port"`
	ListenerGRPCPort int      `mapstructure:"listener-grpc-port"`
	ListenerPipe     string   `mapstructure:"listener-pipe"`
	Transport        string   `mapstructure:"transport"`
	RabbitMQURL      string   `mapstructure:"rabbitmq-url"`
	KafkaBrokers     []string `mapstructure:"kafka-brokers"`
	DriverName       string   `mapstructure:"driver-name"`
	DriverType       string   `mapstructure:"driver-type"`
	Organization     string   `mapstructure:"organization"`
	Tracer           string   `mapstructure:"tracer"`
	Debug            bool     `mapstructure:"debug"`
}

var sidecarCfg sidecarConfig

func init() {
	log.SetLevel(log.DebugLevel)
	viper.SetEnvPrefix("dispatch")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// NewCmd creates a top-level command for event driver sidecar
func NewCmd(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:  "event-driver-sidecar",
		RunE: sidecarCmd,
	}

	cobra.OnInitialize(func() {
		if err := viper.Unmarshal(&sidecarCfg); err != nil {
			panic(errors.Errorf("Error while unmarshalling configuration: %s", err))
		}
		if sidecarCfg.Debug {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	})

	cmds.PersistentFlags().Int("listener-http-port", 8080, "TCP port for HTTP listener to listen on ($DISPATCH_LISTENER_HTTP_PORT)")
	cmds.PersistentFlags().Int("listener-grpc-port", 8081, "TCP port for GRPC listener to listen on ($DISPATCH_LISTENER_GRPC_PORT)")
	cmds.PersistentFlags().String("listener-pipe", "/dispatch-pipe", "The path for named pipe for pipe listener ($DISPATCH_LISTENER_PIPE")
	cmds.PersistentFlags().String("transport", "kafka", "transport backend to use. One of [rabbitmq, kafka, noop] ($DISPATCH_TRANSPORT)")
	cmds.PersistentFlags().String("rabbitmq-url", "amqp://guest:guest@localhost:5672", "If RabbitMQ is used, URL to RABBITMQ Broker ($DISPATCH_RABBITMQ_URL)")
	cmds.PersistentFlags().String("organization", "dispatch", "Organization name to use when routing messages ($DISPATCH_ORGANIZATION)")
	cmds.PersistentFlags().String("driver-name", "", "Name the driver was deployed with. ($DISPATCH_DRIVER_NAME)")
	cmds.PersistentFlags().String("driver-type", "", "Driver type used to deploy this driver. ($DISPATCH_DRIVER_TYPE)")
	cmds.PersistentFlags().String("tracer", "", "OpenTracing-compatible tracer endpoint ($DISPATCH_TRACER)")
	cmds.PersistentFlags().Bool("debug", false, "Debug mode ($DISPATCH_DEBUG)")
	cmds.PersistentFlags().StringSlice("kafka-brokers", []string{"localhost:9092"}, "hostname:port for Kafka broker(s) ($DISPATCH_KAFKA_BROKERS)")

	cmds.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})

	return cmds
}

func sidecarCmd(*cobra.Command, []string) error {

	t, err := createTransport()
	if err != nil {
		return err
	}
	serviceName := fmt.Sprintf("%s-%s", viper.GetString("driver-type"), viper.GetString("driver-name"))
	tracer, closer, err := utils.CreateTracer(serviceName, viper.GetString("tracer"))
	if err != nil {
		return errors.Wrap(err, "error creating a tracer")
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	sharedListener := createSharedListener(t)

	l, err := createListener(sharedListener)
	if err != nil {
		return err
	}

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		<-signals
		l.Shutdown()
	}()

	if err = l.Serve(); err != nil {
		return errors.Wrap(err, "error returned from listener")
	}
	return nil
}

func createListener(sharedListener listener.SharedListener) (EventListener, error) {
	// TODO: create all listeners at once. Currently only HTTP listener supported
	return listener.NewHTTP(sharedListener, sidecarCfg.ListenerHTTPPort)
}

func createTransport() (t events.Transport, err error) {
	switch sidecarCfg.Transport {
	case "kafka":
		t, err = transport.NewKafka(
			sidecarCfg.KafkaBrokers,
			1,
			transport.OptKafkaSendOnly(),
		)
	case "rabbitmq":
		t, err = transport.NewRabbitMQ(
			sidecarCfg.RabbitMQURL,
			transport.OptRabbitMQSendOnly(),
		)
	case "noop":
		t = transport.NewNoop(os.Stdout)
	default:
		panic(fmt.Errorf("transport %s not supported", sidecarCfg.Transport))
	}
	return
}

func createSharedListener(transport events.Transport) listener.SharedListener {
	return listener.NewSharedListener(
		transport,
		&parser.JSONEventParser{},
		validator.NewDefaultValidator(),
		sidecarCfg.Organization,
		sidecarCfg.DriverType,
	)
}
