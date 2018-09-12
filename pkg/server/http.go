///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import "github.com/vmware/dispatch/pkg/http"

func httpServer(config *serverConfig) *http.Server {
	server := http.NewServer(nil)
	server.Host = config.Host
	server.Port = config.Port
	server.TLSPort = config.TLSPort
	server.TLSCertificate = config.TLSCertificate
	server.TLSCertificateKey = config.TLSCertificateKey

	if config.EnableTLS {
		if config.DisableHTTP {
			server.EnabledListeners = []string{"https"}
		} else {
			server.EnabledListeners = []string{"http", "https"}
		}
	}

	return server
}
