///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package parser

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/vmware/dispatch/pkg/events"
)

// JSONEventParser implements events.StreamParser assuming JSON input.
type JSONEventParser struct {
}

// Parse takes io.Reader, and returns slice of CloudEvent. input can be a JSON list, single object,
// or list of JSON objects separated by new line characters
func (p *JSONEventParser) Parse(input io.Reader) ([]events.CloudEvent, error) {
	var e []events.CloudEvent

	reader := bufio.NewReader(input)
	firstChar, err := reader.Peek(1)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(reader)
	if string(firstChar) == "[" {
		// We already have list in our input
		err = decoder.Decode(&e)
	} else {
		for {
			// Parse elements one by one, in case we're dealing with entries separated by new line characters
			var event events.CloudEvent
			err = decoder.Decode(&event)
			if err != nil {
				break
			}
			e = append(e, event)
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return e, nil
}
