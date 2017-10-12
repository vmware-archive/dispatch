//// +build unit

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
    "os"
    "fmt"
    "bufio"
)

const (
    TEST_FILE = "TEST_AUTH_FILE"
    NON_EXISTING_TEST_FILE = "NON_EXISTING_TEST_FILE"
    EXPECTED_OPENWHISK_HOST = "192.168.9.100"
    EXPECTED_OPENWHISK_PORT = "443"
    EXPECTED_OPENWHISK_PRO = "https"
    EXPECTED_TEST_AUTH_KEY = EXPECTED_API_GW_SPACE_SUID + ":123zO3xZCLrMN6v2BKK1dXYFpXlPkccOFqm12CdAsMgRU4VrNZ9lyGVCGuMDGouh"
    EXPECTED_API_HOST= EXPECTED_OPENWHISK_PRO + "://" + EXPECTED_OPENWHISK_HOST + "/api"
    EXPECTED_HOST= EXPECTED_OPENWHISK_HOST + ":" + EXPECTED_OPENWHISK_PORT
    EXPECTED_AUTH_API_KEY = "EXPECTED_AUTH_API_KEY"
    EXPECTED_API_GW_SPACE_SUID = "32kc46b1-71f6-4ed5-8c54-816aa4f8c502"
    APIGW_SPACE_SUID = "APIGW_SPACE_SUID"
    EXPECTED_API_VERSION = "v1"
    EXPECTED_CERT = "EXPECTED_CERT"
    EXPECTED_KEY = "EXPECTED_KEY"

    EXPECTED_API_HOST_WHISK = "localhost"
    EXPECTED_TEST_AUTH_KEY_WHISK = "EXPECTED_TEST_AUTH_KEY_WHISK"
    EXPECTED_NAMESPACE_WHISK = "EXPECTED_NAMESPACE_WHISK"
    EXPECTED_AUTH_API_KEY_WHISK = "EXPECTED_AUTH_API_KEY_WHISK"
    EXPECTED_API_GW_SPACE_SUID_WHISK = "EXPECTED_API_GW_SPACE_SUID_WHISK"
    EXPECTED_API_VERSION_WHISK = "EXPECTED_API_VERSION_WHISK"
    EXPECTED_CERT_WHISK = "EXPECTED_CERT_WHISK"
    EXPECTED_KEY_WHISK = "EXPECTED_KEY_WHISK"

    EXPECTED_API_HOST_LOCAL_CONF = "hostname"
    EXPECTED_TEST_AUTH_KEY_LOCAL_CONF = "EXPECTED_TEST_AUTH_KEY_LOCAL_CONF"
    EXPECTED_NAMESPACE_LOCAL_CONF = "EXPECTED_NAMESPACE_LOCAL_CONF"
    EXPECTED_AUTH_API_KEY_LOCAL_CONF = "EXPECTED_AUTH_API_KEY_LOCAL_CONF"
    EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF = "EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF"
    EXPECTED_API_VERSION_LOCAL_CONF = "EXPECTED_API_VERSION_LOCAL_CONF"
    EXPECTED_CERT_LOCAL_CONF = "EXPECTED_CERT_LOCAL_CONF"
    EXPECTED_KEY_LOCAL_CONF = "EXPECTED_KEY_LOCAL_CONF"

    MISSING_AUTH_MESSAGE = "Authentication key is missing"
    MISSING_URL_MESSAGE = "OpenWhisk API host is missing"
)

type FakeOSPackage struct {
    StoredValues map[string]string
}

func (osPackage FakeOSPackage) Getenv(key string, defaultValue string) string {
    if val, ok := osPackage.StoredValues[key]; ok {
        return val
    } else {
        return defaultValue
    }
}

type FakeViperImp struct {
    StoredValues map[string]string
    ReadInErr error
}

func (viperImp FakeViperImp) GetString(key string, defaultvalue string) string {
    if val, ok := viperImp.StoredValues[key]; ok {
        return val
    } else {
        return defaultvalue
    }
}

func (viperImp FakeViperImp) ReadInConfig() error {
    return viperImp.ReadInErr
}

func (viperImp FakeViperImp) SetConfigName(in string) {
}

func (viperImp FakeViperImp) AddConfigPath(in string) {
}

func getCurrentDir() string {
    dir, err := os.Getwd()
    if err != nil {
        return os.Getenv("GOPATH") + "/src/github.com/apache/incubator-openwhisk-client-go/whisk"
    }
    return dir
}

type FakePropertiesImp struct {
    StoredValues_LOCAL_CONF map[string]string
    StoredValues_WHISK map[string]string
}

func (pi FakePropertiesImp) GetPropsFromWskprops(path string) *Wskprops {
    dep := Wskprops {
        APIHost: GetValue(pi.StoredValues_LOCAL_CONF, APIHOST, ""),
        AuthKey: GetValue(pi.StoredValues_LOCAL_CONF, AUTH, ""),
        Namespace: GetValue(pi.StoredValues_LOCAL_CONF, NAMESPACE, ""),
        AuthAPIGWKey: GetValue(pi.StoredValues_LOCAL_CONF, APIGW_ACCESS_TOKEN, ""),
        APIGWSpaceSuid: GetValue(pi.StoredValues_LOCAL_CONF, APIGW_SPACE_SUID, ""),
        Cert: GetValue(pi.StoredValues_LOCAL_CONF, CERT, ""),
        Key: GetValue(pi.StoredValues_LOCAL_CONF, KEY, ""),
        Apiversion: GetValue(pi.StoredValues_LOCAL_CONF, APIVERSION, ""),
    }

    return &dep
}

func (pi FakePropertiesImp) GetPropsFromWhiskProperties() *Wskprops {
    dep := Wskprops {
        APIHost: pi.StoredValues_WHISK[APIHOST],
        AuthKey: pi.StoredValues_WHISK[AUTH],
        Namespace: pi.StoredValues_WHISK[NAMESPACE],
        AuthAPIGWKey: pi.StoredValues_WHISK[APIGW_ACCESS_TOKEN],
        APIGWSpaceSuid: pi.StoredValues_WHISK[APIGW_SPACE_SUID],
        Cert: pi.StoredValues_WHISK[CERT],
        Key: pi.StoredValues_WHISK[KEY],
        Apiversion: pi.StoredValues_WHISK[APIVERSION],
    }
    return &dep
}

func CreateFile(lines []string, path string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

    w := bufio.NewWriter(file)
    for _, line := range lines {
        fmt.Fprintln(w, line)
    }
    return w.Flush()
}

func DeleteFile(path string) error {
    return os.Remove(path)
}

func TestGetPropsFromWhiskProperties(t *testing.T) {
    lines := []string{ EXPECTED_TEST_AUTH_KEY }
    CreateFile(lines, TEST_FILE)

    fakeOSPackage := FakeOSPackage {
        StoredValues: map[string]string {
            OPENWHISK_HOME: getCurrentDir(),
        },
    }
    pi := PropertiesImp{
        OsPackage: fakeOSPackage,
    }

    dep := pi.GetPropsFromWhiskProperties()
    assert.Equal(t, DEFAULT_NAMESPACE, dep.Namespace)
    assert.Equal(t, "", dep.AuthKey)
    assert.Equal(t, "", dep.AuthAPIGWKey)
    assert.Equal(t, "", dep.APIHost)
    assert.Equal(t, "", dep.APIGWSpaceSuid)
    assert.Equal(t, DEFAULT_VERSION, dep.Apiversion)
    assert.Equal(t, "", dep.Key)
    assert.Equal(t, "", dep.Cert)
    assert.Equal(t, WHISK_PROPERTY, dep.Source)

    lines = []string{ TEST_AUTH_FILE + "=" + TEST_FILE, OPENWHISK_PRO + "=" + EXPECTED_OPENWHISK_PRO,
        OPENWHISK_PORT + "=" + EXPECTED_OPENWHISK_PORT,
        OPENWHISK_HOST + "=" +  EXPECTED_OPENWHISK_HOST,
    }


    CreateFile(lines, OPENWHISK_PROPERTIES)
    pi = PropertiesImp{
        OsPackage: fakeOSPackage,
    }
    dep = pi.GetPropsFromWhiskProperties()
    assert.Equal(t, DEFAULT_NAMESPACE, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY, dep.AuthKey)
    assert.Equal(t, "", dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_OPENWHISK_HOST, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID, dep.APIGWSpaceSuid)
    assert.Equal(t, DEFAULT_VERSION, dep.Apiversion)
    assert.Equal(t, "", dep.Key)
    assert.Equal(t, "", dep.Cert)
    assert.Equal(t, WHISK_PROPERTY, dep.Source)

    DeleteFile(OPENWHISK_PROPERTIES)

    DeleteFile(NON_EXISTING_TEST_FILE)
    lines = []string{ TEST_AUTH_FILE + "=" + NON_EXISTING_TEST_FILE, OPENWHISK_PRO + "=" + EXPECTED_OPENWHISK_PRO,
        OPENWHISK_PORT + "=" + EXPECTED_OPENWHISK_PORT,
        OPENWHISK_HOST + "=" +  EXPECTED_OPENWHISK_HOST}
    CreateFile(lines, OPENWHISK_PROPERTIES)
    pi = PropertiesImp{
        OsPackage: fakeOSPackage,
    }
    dep = pi.GetPropsFromWhiskProperties()
    assert.Equal(t, DEFAULT_NAMESPACE, dep.Namespace)
    assert.Equal(t, "", dep.AuthKey)
    assert.Equal(t, "", dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_OPENWHISK_HOST, dep.APIHost)
    assert.Equal(t, "", dep.APIGWSpaceSuid)
    assert.Equal(t, DEFAULT_VERSION, dep.Apiversion)
    assert.Equal(t, "", dep.Key)
    assert.Equal(t, "", dep.Cert)
    assert.Equal(t, WHISK_PROPERTY, dep.Source)
    DeleteFile(OPENWHISK_PROPERTIES)

    DeleteFile(TEST_FILE)
}

func TestGetPropsFromWskprops(t *testing.T) {
    lines := []string{ APIHOST + "=" + EXPECTED_HOST, AUTH + "=" + EXPECTED_TEST_AUTH_KEY,
        NAMESPACE + "=" + DEFAULT_NAMESPACE,
        APIGW_ACCESS_TOKEN + "=" + EXPECTED_AUTH_API_KEY, APIVERSION + "=" + EXPECTED_API_VERSION,
        KEY + "=" + EXPECTED_KEY, CERT + "=" + EXPECTED_CERT}
    CreateFile(lines, DEFAULT_LOCAL_CONFIG)

    fakeOSPackage := FakeOSPackage{
        StoredValues: map[string]string {
            HOMEPATH: getCurrentDir(),
        },
    }
    pi := PropertiesImp{
        OsPackage: fakeOSPackage,
    }

    dep := pi.GetPropsFromWskprops("")
    assert.Equal(t, DEFAULT_NAMESPACE, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_HOST, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION, dep.Apiversion)
    assert.Equal(t, EXPECTED_KEY, dep.Key)
    assert.Equal(t, EXPECTED_CERT, dep.Cert)
    assert.Equal(t, WSKPROP, dep.Source)

    path := getCurrentDir() + "/" + DEFAULT_LOCAL_CONFIG
    dep = pi.GetPropsFromWskprops(path)
    assert.Equal(t, DEFAULT_NAMESPACE, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_HOST, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION, dep.Apiversion)
    assert.Equal(t, EXPECTED_KEY, dep.Key)
    assert.Equal(t, EXPECTED_CERT, dep.Cert)
    assert.Equal(t, WSKPROP, dep.Source)

    DeleteFile(DEFAULT_LOCAL_CONFIG)
}

func TestGetDefaultConfigFromProperties(t *testing.T) {
    fakeProperties := FakePropertiesImp{
        StoredValues_LOCAL_CONF: map[string]string {
            APIHOST: EXPECTED_OPENWHISK_HOST,
            AUTH: EXPECTED_AUTH_API_KEY,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err := GetDefaultConfigFromProperties(fakeProperties)
    assert.Equal(t, DEFAULT_NAMESPACE, config.Namespace)
    assert.Equal(t, EXPECTED_CERT, config.Cert)
    assert.Equal(t, EXPECTED_KEY, config.Key)
    assert.Equal(t, EXPECTED_AUTH_API_KEY, config.AuthToken)
    assert.Equal(t, EXPECTED_OPENWHISK_HOST, config.Host)
    assert.Equal(t, EXPECTED_API_HOST, config.BaseURL.String())
    assert.Equal(t, EXPECTED_API_VERSION, config.Version)
    assert.False(t, config.Verbose)
    assert.False(t, config.Debug)
    assert.True(t, config.Insecure)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: map[string]string {
            AUTH: EXPECTED_AUTH_API_KEY,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err = GetDefaultConfigFromProperties(fakeProperties)
    assert.Equal(t, DEFAULT_NAMESPACE, config.Namespace)
    assert.Equal(t, EXPECTED_CERT, config.Cert)
    assert.Equal(t, EXPECTED_KEY, config.Key)
    assert.Equal(t, EXPECTED_AUTH_API_KEY, config.AuthToken)
    assert.Equal(t, "", config.Host)
    assert.Nil(t, config.BaseURL)
    assert.Equal(t, EXPECTED_API_VERSION, config.Version)
    assert.False(t, config.Verbose)
    assert.False(t, config.Debug)
    assert.True(t, config.Insecure)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_URL_MESSAGE)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: map[string]string {
            APIHOST: EXPECTED_OPENWHISK_HOST,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err = GetDefaultConfigFromProperties(fakeProperties)
    assert.Equal(t, DEFAULT_NAMESPACE, config.Namespace)
    assert.Equal(t, EXPECTED_CERT, config.Cert)
    assert.Equal(t, EXPECTED_KEY, config.Key)
    assert.Equal(t, "", config.AuthToken)
    assert.Equal(t, EXPECTED_OPENWHISK_HOST, config.Host)
    assert.Equal(t, EXPECTED_API_HOST, config.BaseURL.String())
    assert.Equal(t, EXPECTED_API_VERSION, config.Version)
    assert.False(t, config.Verbose)
    assert.False(t, config.Debug)
    assert.True(t, config.Insecure)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_AUTH_MESSAGE)
}

func TestGetConfigFromWskprops(t *testing.T) {
    fakeProperties := FakePropertiesImp{
        StoredValues_LOCAL_CONF: map[string]string {
            APIHOST: EXPECTED_OPENWHISK_HOST,
            AUTH: EXPECTED_AUTH_API_KEY,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err := GetConfigFromWskprops(fakeProperties, "")
    assert.Equal(t, DEFAULT_NAMESPACE, config.Namespace)
    assert.Equal(t, EXPECTED_CERT, config.Cert)
    assert.Equal(t, EXPECTED_KEY, config.Key)
    assert.Equal(t, EXPECTED_AUTH_API_KEY, config.AuthToken)
    assert.Equal(t, EXPECTED_OPENWHISK_HOST, config.Host)
    assert.Equal(t, EXPECTED_API_HOST, config.BaseURL.String())
    assert.Equal(t, EXPECTED_API_VERSION, config.Version)
    assert.False(t, config.Verbose)
    assert.False(t, config.Debug)
    assert.True(t, config.Insecure)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: map[string]string {
            AUTH: EXPECTED_AUTH_API_KEY,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err = GetConfigFromWskprops(fakeProperties, "")
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_URL_MESSAGE)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: map[string]string {
            APIHOST: EXPECTED_OPENWHISK_HOST,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err = GetConfigFromWskprops(fakeProperties, "")
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_AUTH_MESSAGE)
}

func TestGetConfigFromWhiskProperties(t *testing.T) {
    fakeProperties := FakePropertiesImp{
        StoredValues_WHISK: map[string]string {
            APIHOST: EXPECTED_OPENWHISK_HOST,
            AUTH: EXPECTED_AUTH_API_KEY,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err := GetConfigFromWhiskProperties(fakeProperties)
    assert.Equal(t, DEFAULT_NAMESPACE, config.Namespace)
    assert.Equal(t, EXPECTED_CERT, config.Cert)
    assert.Equal(t, EXPECTED_KEY, config.Key)
    assert.Equal(t, EXPECTED_AUTH_API_KEY, config.AuthToken)
    assert.Equal(t, EXPECTED_OPENWHISK_HOST, config.Host)
    assert.Equal(t, EXPECTED_API_HOST, config.BaseURL.String())
    assert.Equal(t, EXPECTED_API_VERSION, config.Version)
    assert.False(t, config.Verbose)
    assert.False(t, config.Debug)
    assert.True(t, config.Insecure)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_WHISK: map[string]string {
            AUTH: EXPECTED_AUTH_API_KEY,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err = GetConfigFromWhiskProperties(fakeProperties)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_URL_MESSAGE)

    fakeProperties = FakePropertiesImp{
        StoredValues_WHISK: map[string]string {
            APIHOST: EXPECTED_OPENWHISK_HOST,
            NAMESPACE: DEFAULT_NAMESPACE,
            APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY,
            APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID,
            APIVERSION: EXPECTED_API_VERSION,
            CERT: EXPECTED_CERT,
            KEY: EXPECTED_KEY,
        },
    }

    config, err = GetConfigFromWhiskProperties(fakeProperties)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_AUTH_MESSAGE)
}

func TestValidateWskprops(t *testing.T) {
    dep := Wskprops {
        AuthKey: "",
        Namespace: DEFAULT_NAMESPACE,
        AuthAPIGWKey: "",
        APIGWSpaceSuid: "",
        Apiversion: DEFAULT_VERSION,
        Key: "",
        Cert: "",
    }
    err := ValidateWskprops(&dep)
    assert.Contains(t, err.Error(), MISSING_URL_MESSAGE)

    dep = Wskprops {
        APIHost: EXPECTED_OPENWHISK_HOST,
        AuthKey: "",
        Namespace: DEFAULT_NAMESPACE,
        AuthAPIGWKey: "",
        APIGWSpaceSuid: "",
        Apiversion: DEFAULT_VERSION,
        Key: "",
        Cert: "",
    }
    err = ValidateWskprops(&dep)
    assert.Contains(t, err.Error(), MISSING_AUTH_MESSAGE)

    dep = Wskprops {
        APIHost: EXPECTED_OPENWHISK_HOST,
        AuthKey: "auth_key",
        Namespace: DEFAULT_NAMESPACE,
        AuthAPIGWKey: "",
        APIGWSpaceSuid: "",
        Apiversion: DEFAULT_VERSION,
        Key: "",
        Cert: "",
    }
    err = ValidateWskprops(&dep)
    assert.Equal(t, nil, err)

}

func TestGetDefaultWskProp(t *testing.T) {
    valid_whisk_values := map[string]string {
        APIHOST: EXPECTED_API_HOST_WHISK,
        AUTH: EXPECTED_TEST_AUTH_KEY_WHISK,
        NAMESPACE: EXPECTED_NAMESPACE_WHISK,
        APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY_WHISK,
        APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID_WHISK,
        APIVERSION: EXPECTED_API_VERSION_WHISK,
        CERT: EXPECTED_CERT_WHISK,
        KEY: EXPECTED_KEY_WHISK,
    }
    valid_local_conf_values := map[string]string {
        APIHOST: EXPECTED_API_HOST_LOCAL_CONF,
        AUTH: EXPECTED_TEST_AUTH_KEY_LOCAL_CONF,
        NAMESPACE: EXPECTED_NAMESPACE_LOCAL_CONF,
        APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY_LOCAL_CONF,
        APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF,
        APIVERSION: EXPECTED_API_VERSION_LOCAL_CONF,
        CERT: EXPECTED_CERT_LOCAL_CONF,
        KEY: EXPECTED_KEY_LOCAL_CONF,
    }

    missing_auth_local_conf_values := map[string]string{}
    for k,v := range valid_local_conf_values {
        if k != AUTH {
            missing_auth_local_conf_values[k] = v
        }
    }

    missing_url_local_conf_values := map[string]string{}
    for k,v := range valid_local_conf_values {
        if k != APIHOST {
            missing_url_local_conf_values[k] = v
        }
    }

    missing_auth_whisk_values := map[string]string{}
    for k,v := range valid_whisk_values {
        if k != AUTH {
            missing_auth_whisk_values[k] = v
        }
    }

    missing_url_whisk_values := map[string]string{}
    for k,v := range valid_whisk_values {
        if k != APIHOST {
            missing_url_whisk_values[k] = v
        }
    }

    fakeProperties := FakePropertiesImp{
        StoredValues_WHISK: valid_whisk_values,
    }
    dep, err := GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_WHISK, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_WHISK, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_WHISK, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_WHISK, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_WHISK, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_WHISK, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_WHISK, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_WHISK, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: valid_local_conf_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_LOCAL_CONF, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_LOCAL_CONF, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: valid_local_conf_values,
        StoredValues_WHISK: valid_whisk_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_LOCAL_CONF, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_LOCAL_CONF, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_url_local_conf_values,
        StoredValues_WHISK: valid_whisk_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_WHISK, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_WHISK, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_WHISK, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_WHISK, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_WHISK, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_WHISK, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_WHISK, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_WHISK, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_auth_local_conf_values,
        StoredValues_WHISK: valid_whisk_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_WHISK, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_WHISK, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_WHISK, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_WHISK, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_WHISK, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_WHISK, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_WHISK, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_WHISK, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_auth_local_conf_values,
        StoredValues_WHISK: missing_auth_whisk_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, "", dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_LOCAL_CONF, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.NotEqual(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_auth_local_conf_values,
        StoredValues_WHISK: missing_url_whisk_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, "", dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_LOCAL_CONF, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.NotEqual(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_url_local_conf_values,
        StoredValues_WHISK: missing_auth_whisk_values,
    }
    dep, err = GetDefaultWskProp(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_LOCAL_CONF, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, "", dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.NotEqual(t, nil, err)

}

func TestGetWskPropFromWskprops(t *testing.T) {
    valid_local_conf_values := map[string]string{
        APIHOST: EXPECTED_API_HOST_LOCAL_CONF,
        AUTH: EXPECTED_TEST_AUTH_KEY_LOCAL_CONF,
        NAMESPACE: EXPECTED_NAMESPACE_LOCAL_CONF,
        APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY_LOCAL_CONF,
        APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF,
        APIVERSION: EXPECTED_API_VERSION_LOCAL_CONF,
        CERT: EXPECTED_CERT_LOCAL_CONF,
        KEY: EXPECTED_KEY_LOCAL_CONF,
    }

    missing_auth_local_conf_values := map[string]string{}
    for k, v := range valid_local_conf_values {
        if k != AUTH {
            missing_auth_local_conf_values[k] = v
        }
    }

    missing_url_local_conf_values := map[string]string{}
    for k, v := range valid_local_conf_values {
        if k != APIHOST {
            missing_url_local_conf_values[k] = v
        }
    }

    fakeProperties := FakePropertiesImp{
        StoredValues_LOCAL_CONF: valid_local_conf_values,
    }

    dep, err := GetWskPropFromWskprops(fakeProperties, "")
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_LOCAL_CONF, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_LOCAL_CONF, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_url_local_conf_values,
    }
    dep, err = GetWskPropFromWskprops(fakeProperties, "")
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_LOCAL_CONF, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, "", dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_URL_MESSAGE)

    fakeProperties = FakePropertiesImp{
        StoredValues_LOCAL_CONF: missing_auth_local_conf_values,
    }
    dep, err = GetWskPropFromWskprops(fakeProperties, "")
    assert.Equal(t, EXPECTED_NAMESPACE_LOCAL_CONF, dep.Namespace)
    assert.Equal(t, "", dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_LOCAL_CONF, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_LOCAL_CONF, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_LOCAL_CONF, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_LOCAL_CONF, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_LOCAL_CONF, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_LOCAL_CONF, dep.Key)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_AUTH_MESSAGE)
}

func TestGetWskPropFromWhiskProperty(t *testing.T) {
    valid_whisk_values := map[string]string {
        APIHOST: EXPECTED_API_HOST_WHISK,
        AUTH: EXPECTED_TEST_AUTH_KEY_WHISK,
        NAMESPACE: EXPECTED_NAMESPACE_WHISK,
        APIGW_ACCESS_TOKEN: EXPECTED_AUTH_API_KEY_WHISK,
        APIGW_SPACE_SUID: EXPECTED_API_GW_SPACE_SUID_WHISK,
        APIVERSION: EXPECTED_API_VERSION_WHISK,
        CERT: EXPECTED_CERT_WHISK,
        KEY: EXPECTED_KEY_WHISK,
    }

    missing_auth_whisk_values := map[string]string{}
    for k,v := range valid_whisk_values {
        if k != AUTH {
            missing_auth_whisk_values[k] = v
        }
    }

    missing_url_whisk_values := map[string]string{}
    for k,v := range valid_whisk_values {
        if k != APIHOST {
            missing_url_whisk_values[k] = v
        }
    }

    fakeProperties := FakePropertiesImp{
        StoredValues_WHISK: valid_whisk_values,
    }

    dep, err := GetWskPropFromWhiskProperty(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_WHISK, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_WHISK, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_WHISK, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_WHISK, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_WHISK, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_WHISK, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_WHISK, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_WHISK, dep.Key)
    assert.Equal(t, nil, err)

    fakeProperties = FakePropertiesImp{
        StoredValues_WHISK: missing_auth_whisk_values,
    }

    dep, err = GetWskPropFromWhiskProperty(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_WHISK, dep.Namespace)
    assert.Equal(t, "", dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_WHISK, dep.AuthAPIGWKey)
    assert.Equal(t, EXPECTED_API_HOST_WHISK, dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_WHISK, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_WHISK, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_WHISK, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_WHISK, dep.Key)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_AUTH_MESSAGE)

    fakeProperties = FakePropertiesImp{
        StoredValues_WHISK: missing_url_whisk_values,
    }

    dep, err = GetWskPropFromWhiskProperty(fakeProperties)
    assert.Equal(t, EXPECTED_NAMESPACE_WHISK, dep.Namespace)
    assert.Equal(t, EXPECTED_TEST_AUTH_KEY_WHISK, dep.AuthKey)
    assert.Equal(t, EXPECTED_AUTH_API_KEY_WHISK, dep.AuthAPIGWKey)
    assert.Equal(t, "", dep.APIHost)
    assert.Equal(t, EXPECTED_API_GW_SPACE_SUID_WHISK, dep.APIGWSpaceSuid)
    assert.Equal(t, EXPECTED_API_VERSION_WHISK, dep.Apiversion)
    assert.Equal(t, EXPECTED_CERT_WHISK, dep.Cert)
    assert.Equal(t, EXPECTED_KEY_WHISK, dep.Key)
    assert.NotEqual(t, nil, err)
    assert.Contains(t, err.Error(), MISSING_URL_MESSAGE)

}
