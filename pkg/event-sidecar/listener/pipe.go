///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package listener

import (
	"bufio"
	"errors"
	"log"
	"os"
	"syscall"

	"github.com/vmware/dispatch/pkg/events"
)

// Event sidecar pipe files
const (
	EventPipe    = "/pipes/events"
	ResponsePipe = "/pipes/response"
)

// PipeListener type for event pipe channels
type PipeListener struct {
}

// NewPipe creates a new PipeListener - TODO: finish me
func NewPipe() (*PipeListener, error) {
	log.Print("Creating new pipe listener")
	err := syscall.Mkfifo(EventPipe, 0666)
	if err != nil {
		log.Printf("error creating input pipe: %v", err)
		err = os.Remove(EventPipe)
		if err != nil {
			log.Printf("error removing input pipe: %v", err)
			return nil, nil
		}
		err = syscall.Mkfifo(EventPipe, 0666)
		if err != nil {
			return nil, nil
		}
	} else {
		log.Printf("Created %v\n", EventPipe)
	}

	infile, err := os.OpenFile(EventPipe, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(infile)

	go func() {
		for {
			_, err := readMessage(reader)
			if err != nil {
				log.Printf("error reading message: %v", err)
				continue
			}
		}
	}()

	return nil, nil
}

func readMessage(reader *bufio.Reader) (*events.CloudEvent, error) {
	return nil, errors.New("not implemented")
}
