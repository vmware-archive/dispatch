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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/vmware/dispatch/pkg/event-sidecar/listener"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/parser"
	"github.com/vmware/dispatch/pkg/events/transport"
	"github.com/vmware/dispatch/pkg/events/validator"
)

// NO TESTS

type sidecarConfig struct {
	ListenerProtocol string `mapstructure:"listener-protocol"`
	ListenerPort     int    `mapstructure:"listener-port"`
	ListenerPipe     string `mapstructure:"listener-pipe"`
	Transport        string `mapstructure:"transport"`
	RabbitMQURL      string `mapstructure:"rabbitmq-url"`
	Tenant           string `mapstructure:"tenant"`
	TracerURL        string `mapstructure:"tracer-url"`
	Debug            bool   `mapstructure:"debug"`
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
		Use: "event-driver-sidecar",
		Run: sidecarCmd,
	}

	cobra.OnInitialize(func() {
		if err := viper.Unmarshal(&sidecarCfg); err != nil {
			log.Fatalf("Error while unmarshalling configuration: %s", err)
		}
		if sidecarCfg.Debug {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	})

	cmds.PersistentFlags().String("listener-protocol", "http", "Protocol to be used by listener. One of [http, grpc, pipe] ($DISPATCH_LISTENER_PROTOCOL)")
	cmds.PersistentFlags().Int("listener-port", 8080, "If http or grpc listener is used, specify the TCP port to listen on ($DISPATCH_LISTENER_PORT)")
	cmds.PersistentFlags().String("listener-pipe", "/dispatch-pipe", "If pipe listener is used, specify the path for named pipe ($DISPATCH_LISTENER_PIPE")
	cmds.PersistentFlags().String("transport", "rabbitmq", "transport backend to use. One of [rabbitmq, kafka, noop] ($DISPATCH_TRANSPORT)")
	cmds.PersistentFlags().String("rabbitmq-url", "amqp://guest:guest@localhost:5672", "If RabbitMQ is used, URL to RABBITMQ Broker ($DISPATCH_RABBITMQ_URL)")
	cmds.PersistentFlags().String("tenant", "dispatch", "Tenant name to use when routing messages ($DISPATCH_TENANT)")
	cmds.PersistentFlags().String("tracer-url", "", "URL to OpenTracing-compatible tracer ($DISPATCH_TRACER_URL)")
	cmds.PersistentFlags().Bool("debug", false, "Debug mode ($DISPATCH_DEBUG)")

	cmds.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})

	return cmds
}

func sidecarCmd(*cobra.Command, []string) {
	t, err := createTransport()
	if err != nil {
		log.Fatalln(err)
	}

	sharedListener := createSharedListener(t)

	l, err := createListener(sharedListener)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		<-signals
		l.Shutdown()
	}()

	if err = l.Serve(); err != nil {
		log.Fatalf("Error returned from listener: %s\n", err)
	}
}

func createListener(sharedListener listener.SharedListener) (l EventListener, err error) {
	switch sidecarCfg.ListenerProtocol {
	case "http":
		l, err = listener.NewHTTP(sharedListener, sidecarCfg.ListenerPort)
		// TODO: add support for pipe & gRPC
	default:
		panic(fmt.Errorf("protocol %s not supported", sidecarCfg.ListenerProtocol))
	}
	return
}

func createTransport() (t events.Transport, err error) {
	switch sidecarCfg.Transport {
	case "rabbitmq":
		t, err = transport.NewRabbitMQ(
			sidecarCfg.RabbitMQURL,
			sidecarCfg.Tenant,
		)
	case "noop":
		t = transport.NewNoop(os.Stdout)
		// TODO: add support for Kafka
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
		sidecarCfg.Tenant,
	)
}
