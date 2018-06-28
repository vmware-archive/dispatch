///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"testing"
)

func Test_CamelCaseToLowerSeparated(t *testing.T) {
	type camelCaseToLowerSeparatedCase struct {
		given    string
		expected string
	}

	cases := []camelCaseToLowerSeparatedCase{
		{"foo", "foo"},
		{"BAR", "bar"},
		{"Bar", "bar"},
		{"fooBar", "foo.bar"},
		{"FOOBAR", "foobar"},
		{"FOOBar", "foo.bar"},
		{"FooBar日本語", "foo.bar日本語"},
		{"FooBarF日本語", "foo.bar.f日本語"},
		{"FooBarF日本語aaa", "foo.bar.f日本語aaa"},
		{"FooBarF日本語Aaa", "foo.bar.f日本語.aaa"},
		{"FooBarF日本語bazFoz", "foo.bar.f日本語baz.foz"},
		{"", ""},
		{"F00", "f00"},
	}

	for _, c := range cases {
		result := CamelCaseToLowerSeparated(c.given, ".")

		if result != c.expected {
			t.Errorf("given: %s, expected: %s, received: %s", c.given, c.expected, result)
		}
	}
}

func Test_SeparatedToCamelCase(t *testing.T) {
	type separatedToCamelCaseCase struct {
		given    string
		expected string
	}
	cases := []separatedToCamelCaseCase{
		{"foo", "foo"},
		{"BAR", "bar"},
		{"Bar", "bar"},
		{"foo.bar", "fooBar"},
		{"foo.bar日本語", "fooBar日本語"},
		{"", ""},
		{"F00", "f00"},
		{"foo-bar", "foo-bar"},
	}

	for _, c := range cases {
		result := SeparatedToCamelCase(c.given, ".")

		if result != c.expected {
			t.Errorf("given: %s, expected: %s, received: %s", c.given, c.expected, result)
		}
	}
}
