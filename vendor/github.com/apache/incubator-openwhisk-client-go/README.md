# Openwhisk Client Go
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Build Status](https://travis-ci.org/apache/incubator-openwhisk-client-go.svg?branch=master)](https://travis-ci.org/apache/incubator-openwhisk-client-go)

This project `openwhisk-client-go` is a Go client library to access Openwhisk API.


# Disclaimer

This project is currently on an experimental stage. We periodically synchronize the source code of this repository with
the [Go whisk folder](https://github.com/apache/incubator-openwhisk/tree/master/tools/cli/go-whisk) in OpenWhisk. The framework of test cases is under construction
for this repository. Please contribute to the [Go whisk folder](https://github.com/apache/incubator-openwhisk/tree/master/tools/cli/go-whisk) in OpenWhisk for any Go whisk changes, before we officially announce the separation
of OpenWhisk CLI from OpenWhisk.


### Usage

```go
import "github.com/apache/incubator-openwhisk-client-go/whisk"
```

Construct a new whisk client, then use various services to access different parts of the whisk api.  For example to get the `hello` action:

```go
client, _ := whisk.NewClient(http.DefaultClient, nil)
action, resp, err := client.Actions.List("hello")
```

Some API methods have optional parameters that can be passed. For example, to list the first 30 actions, after the 30th action:
```go
client, _ := whisk.NewClient(http.DefaultClient, nil)

options := &whisk.ActionListOptions{
  Limit: 30,
  Skip: 30,
}

actions, resp, err := client.Actions.List(options)
```

Whisk can be configured by passing in a `*whisk.Config` object as the second argument to `whisk.New( ... )`.  For example:

```go
u, _ := url.Parse("https://whisk.stage1.ng.bluemix.net:443/api/v1/")
config := &whisk.Config{
  Namespace: "_",
  AuthKey: "aaaaa-bbbbb-ccccc-ddddd-eeeee",
  BaseURL: u
}
client, err := whisk.Newclient(http.DefaultClient, config)
```


### Example
```go
import (
  "net/http"
  "net/url"

  "github.com/apache/incubator-openwhisk-client-go/whisk"
)

func main() {
  client, err := whisk.NewClient(http.DefaultClient, nil)
  if err != nil {
    fmt.Println(err)
    os.Exit(-1)
  }

  options := &whisk.ActionListOptions{
    Limit: 30,
    Skip: 30,
  }

  actions, resp, err := client.Actions.List(options)
  if err != nil {
    fmt.Println(err)
    os.Exit(-1)
  }

  fmt.Println("Returned with status: ", resp.Status)
  fmt.Println("Returned actions: \n%+v", actions)

}


```
