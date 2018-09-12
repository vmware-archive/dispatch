![Dispatch](docs/assets/images/logo-large.png "Dispatch Logo")

> **NOTE:** This is the knative branch of Dispatch.  Full dispatch functionality is still a ways off.  The code here
> represents a work in progress.  Links to documentation are likely outdated.

Dispatch is a framework for deploying and managing serverless style applications.  The intent is a framework
which enables developers to build applications which are defined by functions which handle business logic and services
which provide all other functionality:

* State (Databases)
* Messaging/Eventing (Queues)
* Ingress (Api-Gateways)
* Etc.

Our goal is to provide a substrate which can be built upon and extended to serve as a framework for serverless
applications.  Additionally, the framework must provide tools and features which aid the developer in building,
debugging and maintaining their serverless application.

## Documentation

Checkout the detailed [documentation](https://vmware.github.io/dispatch) including a [quickstart guide](https://vmware.github.io/dispatch/documentation/guides/quickstart).

## Architecture

The diagram below illustrates the different components which make up the Dispatch project:

![initial dispatch architecture diagram](docs/_specs/dispatch-v2-architecture.png "Initial Architecture")

## Installation

Installing Dispatch depends on having a Kubernetes cluster with the Knative components installed (Build, Serving and soon Eventing).  From here build and install dispatch as follows:

1. Set the following environment variables:
    ```bash
    export DOCKER_URL=https://index.docker.io/v1/
    export DOCKER_REPOSITORY="username or repository"
    export DOCKER_USERNAME="username"
    export DOCKER_PASSWORD="password"
    export DISPATCH_NAMESPACE="dispatch-server"
    export RELEASE_NAME="dispatch-server"
    ```

2. Build and publish a dispatch image:
    ```bash
    PUSH_IMAGES=1 make images
    ```

3. The previous command will output a configuration file `values.yaml`:
    ```yaml
    image:
      host: username
      tag: v0.1.xx
    registry:
      insecure: false
      # Use https://index.docker.io/v1/ for dockerhub
      url: https://index.docker.io/v1/
      repository: repository
      username: ********
      password: ********
    ```

4. Deploy via helm chart (if helm is not installed and initialized, do that first):
    ```bash
    helm upgrade -i --debug --name ${RELEASE_NAME} ./dispatch --namespace ${DISPATCH_NAMESPACE}
    ```

5. Build the CLI (substitute darwin for linux if needed):
    ```bash
    make cli-darwin
    ```

6. Create the Dispatch config:
    ```bash
    cat << EOF > config.json
    {
      "current": "${RELEASE_NAME}",
      "contexts": {
        "${RELEASE_NAME}": {
          "host": "$(kubectl -n ${DISPATCH_NAMESPACE} get service ${RELEASE_NAME}-nginx-ingress-controller -o json | jq -r .status.loadBalancer.ingress[].ip)",
          "port": 443,
          "scheme": "https",
          "insecure": true
        }
      }
    }
    EOF
    ```

7. Test out your install:
    ```bash
    ./dispatch-darwin --config ./config.json create function --image dispatchframework/python3-base:0.0.13-knative hello ./examples/python3/hello.py
    Created function: hello
    ```
    Once status is READY
    ```bash
    ./dispatch-darwin --config ./config.json get function
       NAME  | FUNCTIONIMAGE | STATUS |         CREATED DATE
    ----------------------------------------------------------------
      hello  | ************* | READY  | Thu Sep 13 12:41:07 PDT 2018
    ```
    Exec the function
    ```
    ./dispatch-darwin --config ./config.json exec hello <<< '{"name": "user"}' | jq .
    {
      "context": {
        "logs": {
          "stdout": [
            "messages to stdout show up in logs"
          ],
          "stderr": null
        }
      },
      "payload": {
        "myField": "Hello, user from Nowhere"
      }
    }
    ```

For a more complete quickstart see the [developer documentation](#documentation)
