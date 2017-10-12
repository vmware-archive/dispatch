// +build unit

/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package whisk

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "net/http"
    "fmt"
    "net/url"
)

const (
    FakeHost = "myUrl.com"
    FakeHostDiff = "myUrlTest.com"
    FakeBaseURL = "https://" + FakeHost + "/api"
    FakeBaseURLDiff = "https://" + FakeHostDiff + "/api"
    FakeNamespace = "my_namespace"
    FakeAuthKey = "dhajfhshfs:hjhfsjfdjfjsgfjs"
)

func GetValidConfigTest() *Config {
    var config Config
    config.Host = FakeHost
    config.Namespace = FakeNamespace
    config.AuthToken = FakeAuthKey
    return &config
}

func GetInvalidConfigMissingApiHostTest() *Config {
    var config Config
    config.Namespace = FakeNamespace
    config.AuthToken = FakeAuthKey
    return &config
}

func GetInvalidConfigMissingApiHostWithBaseURLTest() *Config {
    var config Config
    urlBase := fmt.Sprintf("https://%s/api", FakeHostDiff)
    config.BaseURL, _ = url.Parse(urlBase)
    config.Namespace = FakeNamespace
    config.AuthToken = FakeAuthKey
    return &config
}

func GetValidConfigDiffApiHostAndBaseURLTest() *Config {
    var config Config
    urlBase := fmt.Sprintf("https://%s/api", FakeHostDiff)
    config.BaseURL, _ = url.Parse(urlBase)
    config.Host = FakeHost
    config.Namespace = FakeNamespace
    config.AuthToken = FakeAuthKey
    return &config
}

func TestNewClient(t *testing.T) {
    // Test the use case to pass a valid config.
    config := GetValidConfigTest()
    client, err := NewClient(http.DefaultClient, config)
    assert.Nil(t, err)
    assert.NotNil(t, client)
    assert.Equal(t, FakeNamespace, client.Config.Namespace)
    assert.Equal(t, FakeHost, client.Config.Host)
    assert.Equal(t, FakeBaseURL, client.Config.BaseURL.String())
    assert.Equal(t, FakeAuthKey, client.Config.AuthToken)

    // Test the use case to pass an invalid config with a missing api host.
    config = GetInvalidConfigMissingApiHostTest()
    client, err = NewClient(http.DefaultClient, config)
    assert.NotNil(t, err)
    assert.Contains(t, err.Error(), "Unable to create request URL, because OpenWhisk API host is missing")
    assert.Nil(t, client)

    // Test the use case to pass a valid config with the base url but without api host.
    config = GetInvalidConfigMissingApiHostWithBaseURLTest()
    client, err = NewClient(http.DefaultClient, config)
    assert.NotNil(t, err)
    assert.Contains(t, err.Error(), "Unable to create request URL, because OpenWhisk API host is missing")
    assert.Nil(t, client)

    // Test the use case to pass a valid config with both the base and api host of different values.
    config = GetValidConfigDiffApiHostAndBaseURLTest()
    client, err = NewClient(http.DefaultClient, config)
    assert.Nil(t, err)
    assert.NotNil(t, client)
    assert.Equal(t, FakeNamespace, client.Config.Namespace)
    assert.Equal(t, FakeHost, client.Config.Host)
    assert.Equal(t, FakeBaseURLDiff, client.Config.BaseURL.String())
    assert.Equal(t, FakeAuthKey, client.Config.AuthToken)
}
