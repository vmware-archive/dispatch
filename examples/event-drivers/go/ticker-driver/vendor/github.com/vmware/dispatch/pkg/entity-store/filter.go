///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entitystore

const (
	// FilterVerbIn tests containment
	FilterVerbIn Verb = "in"

	// FilterVerbEqual tests equality
	FilterVerbEqual Verb = "equal"

	// FilterVerbBefore tests two time.Time
	FilterVerbBefore Verb = "before"

	// FilterVerbAfter tests two time.Time
	FilterVerbAfter Verb = "after"

	// FilterScopeField defines that the subject is a BaseEntity field
	FilterScopeField Scope = "field"

	// FilterScopeTag defines that the subject is a BaseEntity Tag
	FilterScopeTag Scope = "tag"

	// FilterScopeExtra defines that the subject is a extra field extended from BaseEntity
	FilterScopeExtra Scope = "extra"
)

// Verb describe the filter verb
type Verb string

// Scope describes which scope this filter is applied on
type Scope string

// Filter defines a set of criteria to filter entities when listing
type Filter interface {
	Add(...FilterStat) Filter
	FilterStats() []FilterStat
}

// FilterStat (Filter Statement) defines one filter criterion
type FilterStat struct {
	Scope   Scope
	Subject string
	Verb    Verb
	Object  interface{}
}

// FilterStatByApplication defines one filter statement by application name
func FilterStatByApplication(name string) FilterStat {
	return FilterStat{
		Scope:   FilterScopeTag,
		Subject: "Application",
		Verb:    FilterVerbEqual,
		Object:  name,
	}
}

type filter struct {
	statements []FilterStat
}

// FilterEverything creates a filter, which will matches all entities
func FilterEverything() Filter {
	return &filter{}
}

// FilterExists creates a filter, which will filter entities with Delete=true
func FilterExists() Filter {
	f := &filter{}
	f.Add(FilterStat{
		Scope:   FilterScopeField,
		Subject: "Delete",
		Verb:    FilterVerbEqual,
		Object:  false,
	})
	return f
}

// FilterByApplication creates a filter, which will filter based on the supplied application name
func FilterByApplication(app string) Filter {
	f := &filter{}
	f.Add(FilterStatByApplication(app))
	return f
}

func (f *filter) Add(stats ...FilterStat) Filter {
	for _, stat := range stats {
		f.statements = append(f.statements, stat)
	}
	return f
}

func (f filter) FilterStats() []FilterStat {
	return f.statements
}
