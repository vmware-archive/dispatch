///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package utils

import (
	"strings"
	"unicode"
)

// CamelCaseToLowerSeparated converts a camel cased string to a multi-word string delimited by the specified separator
func CamelCaseToLowerSeparated(src string, sep string) string {
	var words []string
	var word []rune
	var last rune
	for _, r := range src {
		if unicode.IsUpper(r) {
			if unicode.IsUpper(last) {
				// We have two uppercase letters in a row, it might be uppercase word like ID or SDK
				word = append(word, r)
			} else {
				// We have uppercase after lowercase, which always means start of a new word
				if len(word) > 0 {
					words = append(words, strings.ToLower(string(word)))
				}
				word = []rune{r}
			}
		} else {
			if unicode.IsUpper(last) && len(word) >= 2 {
				// We have a multi-uppercase word followed by an another word, e.g. "SDKToString",
				// but word variable contains "SDKT". We need to extract "T" as a first letter of a new word
				words = append(words, strings.ToLower(string(word[:len(word)-1])))
				word = []rune{last, r}
			} else {
				word = append(word, r)
			}
		}
		last = r
	}
	if len(word) > 0 {
		words = append(words, strings.ToLower(string(word)))
	}
	return strings.Join(words, sep)
}

// SeparatedToCamelCase converts a multi-word string delimited by a separator to camel cased string.
// Note:- SeparatedToCamelCase does not inverse the result of CamelCaseToLowerSeparated, it's a lossy operation.
func SeparatedToCamelCase(src string, sep string) string {
	words := strings.Split(src, sep)
	words[0] = strings.ToLower(words[0])
	for i, w := range words[1:] {
		words[i+1] = strings.Title(w)
	}
	return strings.Join(words, "")
}
