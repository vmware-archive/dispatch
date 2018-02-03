///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package kong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	ewrapper "github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	jsonContentType       = "application/json"
	urlencodedContentType = "application/x-www-form-urlencoded"
)

// Config represents a configure for Kong Client
type Config struct {
	Host     string
	Upstream string
}

// Client is a http client connecting to a Kong server
type Client struct {
	host       string
	upstream   string
	httpClient *http.Client
}

// API is a struct for Kong API
type API struct {

	// id and created_at is required to update an kong API
	ID        string `json:"id,omitempty"`
	CreatedAt int    `json:"created_at,omitempty"`

	Name        string   `json:"name"`
	UpstreamURL string   `json:"upstream_url,omitempty"`
	URIs        []string `json:"uris,omitempty"`
	Hosts       []string `json:"hosts,omitempty"`
	Methods     []string `json:"methods,omitempty"`
	HTTPSOnly   bool     `json:"https_only,omitempty"`
}

// Plugin is a struct for Kong Plugin
type Plugin struct {
	Name    string                 `json:"name"`
	ID      string                 `json:"id,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Enabled bool                   `json:"enabled,omitempty"`
}

// NewClient creates a new Kong Client
func NewClient(config *Config) (*Client, error) {

	client := &Client{
		host:       config.Host,
		upstream:   config.Upstream,
		httpClient: http.DefaultClient,
	}
	return client, nil
}

// Initialize install neccessary plugins into kong at the begining
func (k *Client) Initialize() error {
	defer trace.Trace("")()

	dispatchTransformer := Plugin{
		Name: "dispatch-transformer",
		Config: map[string]interface{}{
			"config.substitute.input":            "input",
			"config.substitute.output":           "output",
			"config.enable.input":                true,
			"config.enable.output":               true,
			"config.http_method":                 "POST",
			"config.add.header":                  "cookie:cookie",
			"config.header_prefix_for_insertion": "x-dispatch-",
			"config.insert_to_body.header":       "blocking:true",
		},
	}
	err := k.updatePluginByName("", dispatchTransformer.Name, &dispatchTransformer)
	if err != nil {
		return err
	}
	return nil
}

func (k *Client) apiEntityToKong(entity *gateway.API) *API {

	upstream := fmt.Sprintf("http://%s/v1/function/%s/runs", k.upstream, entity.Function)
	a := API{
		ID:          entity.ID,
		CreatedAt:   entity.CreatedAt,
		Name:        entity.Name,
		UpstreamURL: upstream,
		Hosts:       entity.Hosts,
		URIs:        entity.URIs,
		Methods:     entity.Methods,
	}
	if len(entity.Protocols) == 1 && entity.Protocols[0] == "https" {
		a.HTTPSOnly = true
	} else {
		a.HTTPSOnly = false
	}
	if entity.CORS == true {
		// note: OPTIONS is a CORS preflight request
		// it is added by dispatch automatically
		// users should not add them mannually
		a.Methods = append(a.Methods, "OPTIONS")
	}
	return &a
}

func (k *Client) apiKongToEntity(apiKong *API) *gateway.API {

	api := &gateway.API{
		ID:        apiKong.ID,
		CreatedAt: apiKong.CreatedAt,
		Name:      apiKong.Name,
		Hosts:     apiKong.Hosts,
		URIs:      apiKong.URIs,
		Methods:   apiKong.Methods,
	}
	return api
}

func getKongError(function string, resp *http.Response) error {

	bytesOut, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		err = ewrapper.Errorf("kong.%s: %v, %s", function, resp.StatusCode, string(bytesOut))
	} else {
		log.Debugf("error read response body")
		err = ewrapper.Errorf("kong.%s: %v", function, resp.StatusCode)
	}
	return err
}

func (k *Client) getResponse(resp *http.Response, object interface{}) error {
	trace.Trace("")()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = ewrapper.Wrap(err, "error reading response")
		log.Error(err)
		return &errors.DriverError{Err: err}
	}
	err = json.Unmarshal(body, object)
	if err != nil {
		err = ewrapper.Wrap(err, "error unmarshal response")
		log.Error(err)
		return &errors.ObjectMarshalError{Err: err}
	}
	return nil
}

func (k *Client) getAPIFromResponseBody(resp *http.Response) (*gateway.API, error) {
	trace.Trace("")()
	var api API
	err := k.getResponse(resp, &api)
	if err != nil {
		return nil, err
	}
	return k.apiKongToEntity(&api), nil
}

// GetAPI get an API from Kong
func (k *Client) GetAPI(name string) (*gateway.API, error) {
	defer trace.Tracef("name '%s'", name)()

	url := fmt.Sprintf("%s/apis/%s", k.host, name)
	resp, err := k.request("GET", url, jsonContentType, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("kong.getAPI.%s: status code: %v", name, resp.StatusCode)
	switch resp.StatusCode {
	case 200:
		return k.getAPIFromResponseBody(resp)
	default:
		err = getKongError("getAPI", resp)
		return nil, &errors.ObjectNotFoundError{Err: err}
	}
}

func stringContains(array []string, val string) bool {
	for _, el := range array {
		if el == val {
			return true
		}
	}
	return false
}

// AddAPI add an API in Kong
func (k *Client) AddAPI(entity *gateway.API) (*gateway.API, error) {
	defer trace.Tracef("name '%s'", entity.Name)()

	a := k.apiEntityToKong(entity)

	url := fmt.Sprintf("%s/apis/", k.host)
	body, err := json.Marshal(a)
	if err != nil {
		return nil, &errors.ObjectMarshalError{Err: err}
	}
	resp, err := k.request("POST", url, jsonContentType, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result *gateway.API
	log.Debugf("kong.addAPI.%s: status code: %v", entity.Name, resp.StatusCode)
	switch resp.StatusCode {
	case 201:
		result, err = k.getAPIFromResponseBody(resp)
		if err != nil {
			return nil, err
		}
	default:
		err = getKongError("addAPI", resp)
		return nil, &errors.ObjectNotFoundError{Err: err}
	}

	if entity.CORS == true {
		corsPlugin := Plugin{
			Name: "cors",
			Config: map[string]interface{}{
				// TODO: '*' for now, should be able to configure the origin later
				"config.origins": "*",
				// Workaround: fix https://github.com/vmware/dispatch/issues/174
				// OPTIIONS is not an allowed method in kong cors plugin
				"config.methods": strings.Join(entity.Methods, ","),
			},
		}
		err := k.updatePluginByName(a.Name, corsPlugin.Name, &corsPlugin)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// UpdateAPI updates an API in Kong
func (k *Client) UpdateAPI(name string, entity *gateway.API) (*gateway.API, error) {

	// Note: make sure  ID and CreatedAt are set in entity,
	// Kong requires them, which is not documented
	defer trace.Tracef("name '%s'", name)()

	a := k.apiEntityToKong(entity)

	body, err := json.Marshal(a)
	if err != nil {
		return nil, &errors.ObjectMarshalError{Err: err}
	}

	resp, err := k.request("PUT", fmt.Sprintf("%s/apis", k.host), jsonContentType, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("kong.updateAPI.%s: status code: %v", entity.Name, resp.StatusCode)

	var result *gateway.API
	switch resp.StatusCode {
	case 200, 201:
		result, err = k.getAPIFromResponseBody(resp)
		if err != nil {
			return nil, err
		}
	default:
		err = getKongError("updateAPI", resp)
		return nil, &errors.ObjectNotFoundError{Err: err}
	}

	if entity.CORS {
		// try to update
		corsPlugin := Plugin{
			Name: "cors",
			Config: map[string]interface{}{
				// TODO: '*' for now, should be able to configure the origin later
				"config.origins": "*",
				// Workaround: fix https://github.com/vmware/dispatch/issues/174
				// OPTIIONS is not an allowed method in kong cors plugin
				"config.methods": strings.Join(entity.Methods, ","),
			},
		}
		err := k.updatePluginByName(name, "cors", &corsPlugin)
		if err != nil {
			return nil, err
		}
	} else {
		// try to delete
		err := k.deletePluginByName(name, "cors")
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// DeleteAPI delete an API from Kong
func (k *Client) DeleteAPI(api *gateway.API) error {
	defer trace.Tracef("name '%s'", api.Name)()

	if api.CORS == true {
		// Get Plugin ID
		err := k.deletePluginByName(api.Name, "cors")
		if err != nil {
			return err
		}
	}

	resp, err := k.request("DELETE", fmt.Sprintf("%s/apis/%s", k.host, api.Name), jsonContentType, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debugf("kong.deleteAPI.%s: status code: %v", api.Name, resp.StatusCode)
	switch resp.StatusCode {
	case 204:
		return nil
	case 404:
		return &errors.ObjectNotFoundError{Err: fmt.Errorf("api not found")}
	default:
		err = getKongError("deleteAPI", resp)
		return &errors.DriverError{Err: err}
	}
}

func (k *Client) getPluginURL(api, plugin string) string {

	url := fmt.Sprintf("%s", k.host)
	if api != "" {
		url = fmt.Sprintf("%s/apis/%s", url, api)
	}
	url = fmt.Sprintf("%s/plugins", url)
	if plugin != "" {
		url = fmt.Sprintf("%s/%s", url, plugin)
	}
	return url
}

func (k *Client) getPlugins(apiName, pluginName string) ([]Plugin, error) {

	url := k.getPluginURL(apiName, "")
	if pluginName != "" {
		// special case: need to use http query parameters instead of a path parameter
		url = fmt.Sprintf("%s?name=%s", url, pluginName)
	}
	resp, err := k.request("GET", url, urlencodedContentType, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("kong.getPlugins.%s: status code: %v", apiName, resp.StatusCode)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		respObject := struct {
			Total int      `json:"total"`
			Data  []Plugin `json:"data"`
		}{}
		err = k.getResponse(resp, &respObject)
		if err != nil {
			return nil, err
		}
		if respObject.Total == 0 {
			err = getKongError("getPlugins", resp)
			return nil, &errors.ObjectNotFoundError{Err: err}
		}
		return respObject.Data, nil
	}

	err = getKongError("getPlugins", resp)
	if resp.StatusCode == 404 {
		return nil, &errors.ObjectNotFoundError{Err: err}
	}
	return nil, &errors.DriverError{Err: err}
}

func (k *Client) updatePluginByName(apiName, pluginName string, plugin *Plugin) error {

	plugins, err := k.getPlugins(apiName, pluginName)
	if err != nil {
		if _, ok := err.(*errors.ObjectNotFoundError); ok {
			log.Debugf("kong.updatePluginByName.%s.%s: no such plugins, try to add", apiName, pluginName)
			// continue
		} else {
			return err
		}
	}

	if plugins == nil {
		// add an empty plugin with an emtpy ID
		plugins = []Plugin{Plugin{ID: ""}}
	}

	for _, p := range plugins {
		// should only have one plugin though
		err = k.updatePluginByID(apiName, p.ID, plugin)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *Client) updatePluginByID(apiName, pluginID string, plugin *Plugin) error {

	method := "POST" // doesn't exist, use POST to create
	if pluginID != "" {
		// already existed, use PATCH to update
		method = "PATCH"
	}
	reqURL := k.getPluginURL(apiName, pluginID)
	body := url.Values{}
	for k, v := range plugin.Config {
		body.Add(k, fmt.Sprintf("%v", v))
	}
	body.Add("name", plugin.Name)

	// TODO: test if json request also works
	// body, err := json.Marshal(plugin)
	// if err != nil {
	// 	return &errors.ObjectMarshalError{Err: err}
	// }
	// resp, err := k.request(method, reqURL, urlencodedContentType, bytes.NewReader(body))

	resp, err := k.request(method, reqURL, urlencodedContentType, bytes.NewBufferString(body.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debugf("kong.updatePlugin.%s: status code: %v", plugin.ID, resp.StatusCode)
	switch resp.StatusCode {
	case 200, 201:
		return nil
	default:
		err = getKongError("updatePlugin", resp)
		return &errors.DriverError{Err: err}
	}
}

func (k *Client) deletePluginByName(apiName, pluginName string) error {

	// we have to get the plugin ID first, it is forced by Kong
	plugins, err := k.getPlugins(apiName, pluginName)
	if err != nil {
		if _, ok := err.(*errors.ObjectNotFoundError); ok {
			log.Debugf("kong.deletePluginByName.%s.%s: no such plugins, skip", apiName, pluginName)
			return nil
		}
		return err
	}
	for _, p := range plugins {
		err := k.deletePluginByID(apiName, p.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// 204: delete successfully
// 404: object not found
func (k *Client) deletePluginByID(apiName, pluginID string) error {

	url := k.getPluginURL(apiName, pluginID)
	resp, err := k.request("DELETE", url, jsonContentType, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debugf("kong.deletePlugin.[%s]: status code: %v", url, resp.StatusCode)

	// TODO: double check error code
	switch resp.StatusCode {
	case 204:
		return nil
	case 404:
		return &errors.ObjectNotFoundError{Err: fmt.Errorf("plugin not found")}
	default:
		err = getKongError("deletePlugin", resp)
		return &errors.DriverError{Err: err}
	}
}

func (k *Client) request(method, url, contentType string, body io.Reader) (*http.Response, error) {

	log.Debugf("kong request method=%s url=%s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, &errors.DriverError{Err: err}
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, &errors.DriverError{Err: err}
	}
	return resp, nil
}
