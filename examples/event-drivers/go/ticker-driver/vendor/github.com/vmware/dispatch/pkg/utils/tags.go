///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

// NO TESTS

import (
	"fmt"
	"strings"

	es "github.com/vmware/dispatch/pkg/entity-store"
)

// ParseTags parses tags pass from dispatch client,
// a tag is key-value pair string, separate by an equal symbol
// the only valid format of a tag is "<key>=<value>"
func ParseTags(filter es.Filter, tags []string) (es.Filter, error) {
	if filter == nil {
		filter = es.FilterEverything()
	}
	for _, tag := range tags {
		values := strings.Split(tag, "=")
		if len(values) != 2 {
			return nil, fmt.Errorf("error parsing tag '%s': invalid format", tag)
		}
		key, value := values[0], values[1]
		switch strings.ToLower(key) {
		case "application", "app":
			filter.Add(es.FilterStatByApplication(value))
		default:
			return nil, fmt.Errorf("error parsing tag '%s': key '%s' not exist", tag, key)
		}
	}
	return filter, nil
}
