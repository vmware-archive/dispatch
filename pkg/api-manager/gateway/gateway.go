///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package gateway

// NO TEST

// API represents the metadata of an API
type API struct {
	ID        string `json:"id,omitempty"`
	CreatedAt int    `json:"created_at,omitempty"`

	Name     string `json:"name,omitempty"`
	Function string `json:"function,omitempty"`

	Hosts   []string `json:"hosts,omitempty"`
	URIs    []string `json:"uris,omitempty"`
	Methods []string `json:"methods,omitempty"`

	Authentication string `json:"authentication,omitempty"`

	Enabled bool `json:"enabled,omitempty"`

	// i.e. http https
	Protocols []string `json:"protocols,omitempty"`

	// reference to tls certificates (a dispatch secret name)
	// TODO: will be replaced by SNI objects
	TLS string `json:"tls,omitempty"`

	CORS bool `json:"cors,omitempty"`
}

// Gateway defines interfaces the underlying API Gateway provides
type Gateway interface {
	Initialize() error
	AddAPI(api *API) (*API, error)
	GetAPI(name string) (*API, error)
	UpdateAPI(name string, api *API) (*API, error)
	DeleteAPI(api *API) error
}
