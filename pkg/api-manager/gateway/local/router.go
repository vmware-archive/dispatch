///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
)

// Serve sets the handler and starts the API Gateway HTTP server
func (g *Gateway) Serve() error {
	g.Server.SetHandler(g)
	return g.Server.Serve()
}

// Shutdown gracefully stops the HTTP server.
func (g *Gateway) Shutdown() error {
	return g.Server.Shutdown()
}

// ServeHTTP implements http.Handler interface.
func (g *Gateway) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	api := g.matchAPI(cleanHost(req.Host), req.URL.Path, req.Method)
	if api == nil {
		// No match found
		writeErrorResp(rw, 404, "no API found with those values")
		return
	}
	if len(api.Protocols) == 1 {
		expectedProto := strings.ToUpper(api.Protocols[0])
		// api is set to support single protocol
		proto := "HTTP"
		if req.TLS != nil {
			proto = "HTTPS"
		}
		if expectedProto != proto {
			writeErrorResp(rw, 400, fmt.Sprintf("Please use %s protocol", expectedProto))
		}
	}

	if api.CORS {
		addCORS(rw)
	}

	if req.Method == http.MethodOptions {
		writeEmptyResp(rw, 200)
		return
	}

	input, err := getInput(req)
	if err != nil {
		writeErrorResp(rw, 400, err.Error())
		return
	}

	blocking := true
	if req.Header.Get("x-dispatch-blocking") == "false" {
		blocking = false
	}

	run := v1.Run{
		Blocking:     blocking,
		FunctionName: api.Function,
		Input:        input,
		HTTPContext:  getContext(req, api.Function),
	}
	resp, err := g.fnClient.RunFunction(req.Context(), api.OrganizationID, &run)
	if err != nil {
		if be, ok := err.(client.Error); ok {
			writeErrorResp(rw, be.Code(), be.Message())
		} else {
			writeErrorResp(rw, 400, err.Error())
		}
		return
	}
	if blocking == false || resp.Output == nil {
		writeEmptyResp(rw, 200)
		return
	}

	if str, ok := resp.Output.(string); ok && str == "" {
		writeEmptyResp(rw, 200)
		return
	}
	rw.Header().Add("Content-type", "application/json")
	enc := json.NewEncoder(rw)
	enc.Encode(resp.Output)
}

// matchAPI returns first API that matches host and path and method. It finds only first match.
// It starts looking by checking path, as it's expected to be most common filter to be set.
// local gateway is expected to be only used by single-host deployments, thus host filter is checked last.
func (g *Gateway) matchAPI(host, path, method string) *gateway.API {
	log.Debugf("Matching API for host:%s, path:%s, method:%s", host, path, method)
	g.RLock()
	defer g.RUnlock()
	if apis, ok := g.pathLookup[path]; ok {
		for i := range apis {
			if matchAPIAgainst(apis[i], host, method, "") {
				return apis[i]
			}
		}
	}

	if apis, ok := g.methodLookup[method]; ok {
		for i := range apis {
			if matchAPIAgainst(apis[i], host, "", path) {
				return apis[i]
			}
		}
	}

	if apis, ok := g.hostLookup[host]; ok {
		for i := range apis {
			if matchAPIAgainst(apis[i], "", method, path) {
				return apis[i]
			}
		}
	}

	return nil
}

// matchAPIAgainst takes api and checks if it matches against provided host, method and string.
// if the api has nil slice for particular property, that property will always match regardless of
// the actual value in the request.
func matchAPIAgainst(api *gateway.API, host, method, path string) bool {
	foundHost := matchString(api.Hosts, host)
	foundPath := matchString(api.URIs, path)
	var foundMethod bool
	if method == http.MethodOptions {
		foundMethod = true
	} else {
		foundMethod = matchString(api.Methods, method)
	}

	return foundHost && foundMethod && foundPath
}

// matchString checks if needle string is in set slice. if any of these are zero values,
// it returns true.
func matchString(set []string, needle string) bool {
	if len(set) == 0 || needle == "" {
		// empty set means no filtering
		return true
	}
	for _, elem := range set {
		if elem == needle {
			return true
		}
	}
	return false
}

func addCORS(rw http.ResponseWriter) {
	rw.Header().Add("Access-Control-Allow-Origin", "*")
	rw.Header().Add("Access-Control-Allow-Methods", "*")
	rw.Header().Add("Access-Control-Allow-Headers", "*")
}

// getInput processes the request and generates an input for the function.
func getInput(req *http.Request) (interface{}, error) {
	if req.Method == http.MethodGet {
		return processValues(req.URL.Query()), nil
	}

	if req.ContentLength == 0 {
		return map[string]interface{}{}, nil
	}

	contentType := req.Header.Get("Content-type")

	if isFormContentType(contentType) {
		req.ParseMultipartForm(
			int64(10 << 20), // 10 MB
		)
		return processValues(req.Form), nil
	}

	if isJSONContentType(contentType) {
		var input interface{}
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&input)
		if err != nil {
			return nil, errors.New("request body is not json")
		}
		return input, nil
	}

	if contentType != "" {
		return nil, fmt.Errorf("request body type is not supported: %s", contentType)
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getContext(req *http.Request, funcName string) map[string]interface{} {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return map[string]interface{}{
		"args":            req.URL.RawQuery,
		"request":         req.URL.String(),
		"request_uri":     req.RequestURI,
		"scheme":          scheme,
		"server_protocol": req.Proto,
		"upstream_uri":    funcName,
		"uri":             req.RequestURI,
		"method":          req.Method,
	}
}

// processValues converts url.Values into a simplified map, i.e.
// when there is a single value for given key, the value is assigned directly.
// Otherwise, the value slice is assigned.
func processValues(values url.Values) map[string]interface{} {
	simpleMap := make(map[string]interface{}, len(values))
	for key, value := range values {
		if len(value) == 1 {
			simpleMap[key] = value[0]
		} else {
			simpleMap[key] = value
		}
	}
	return simpleMap
}

func isFormContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		switch t {
		case "application/x-www-form-urlencoded", "multipart/form-data":
			return true
		}
	}
	return false
}

// isJSONContentType checks if provided string is a proper JSON content type
func isJSONContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == "application/json" {
			return true
		}
		if strings.HasPrefix(t, "application/") && strings.HasSuffix(t, "+json") {
			return true
		}
	}
	return false
}

// cleanHost returns host without port, if one is set
func cleanHost(host string) string {
	if host == "" {
		return ""
	}
	return strings.Split(host, ":")[0]
}

func writeErrorResp(rw http.ResponseWriter, code int, msg string) {
	rw.Header().Add("Content-type", "application/json")
	rw.WriteHeader(code)
	resp := struct {
		Message string `json:"message"`
	}{Message: msg}
	enc := json.NewEncoder(rw)
	enc.Encode(&resp)
}

func writeEmptyResp(rw http.ResponseWriter, code int) {
	rw.WriteHeader(code)
	rw.(http.Flusher).Flush()
}
