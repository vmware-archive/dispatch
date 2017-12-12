///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vmware/dispatch/pkg/event-driver"
	"github.com/vmware/dispatch/pkg/events/rabbitmq"

	"github.com/vmware/dispatch/pkg/events"
)

// NO TESTS

var driverConfigPath = ""

func init() {
	log.SetLevel(log.DebugLevel)
}

func initConfig() {
	// Don't forget to read config either from vsConfigPath or from home directory!
	if driverConfigPath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(driverConfigPath)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".event-driver")
	}

	viper.ReadInConfig()
}

// NewEventDriverCmd creates a top-level command for event driver
func NewEventDriverCmd(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cobra.OnInitialize(initConfig)
	cmds := &cobra.Command{
		Use:   "event-driver",
		Short: "",
		Long:  "",
		Run:   runHelp,
	}
	cmds.PersistentFlags().StringVar(&driverConfigPath, "config", "", "config file (default is $HOME/.event-driver)")

	cmds.PersistentFlags().String("amqpurl", "amqp://guest:guest@localhost:5672", "url to AMQP Broker")
	viper.BindPFlag("amqpurl", cmds.PersistentFlags().Lookup("amqpurl"))

	cmds.PersistentFlags().String("exchange", "dispatch", "Exchange name to use")
	viper.BindPFlag("exchange", cmds.PersistentFlags().Lookup("exchange"))

	cmds.AddCommand(NewCmdVCenter(out, errOut))

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func makeQueue() (events.Queue, error) {
	queue, err := rabbitmq.New(viper.GetString("amqpurl"), viper.GetString("exchange"))
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func makeDriver(consumer eventdriver.Consumer) (eventdriver.Driver, error) {
	queue, err := makeQueue()
	if err != nil {
		return nil, err
	}
	return eventdriver.New(queue, consumer)
}
