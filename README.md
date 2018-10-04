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

> **NOTE**: The links below refer to an old verion of dispatch and may not be accurrate.

Checkout the detailed [documentation](https://vmware.github.io/dispatch) including a [quickstart guide](https://vmware.github.io/dispatch/documentation/guides/quickstart).

## Architecture

The diagram below illustrates the different components which make up the Dispatch project:

![initial dispatch architecture diagram](docs/_specs/dispatch-v2-architecture.png "Initial Architecture")

## Installation

> **NOTE**: Dispatch depends on a very recent version of Knative serving.  In order to install Knative, follow the
> [development instructions](https://github.com/knative/serving/blob/master/DEVELOPMENT.md)

Installing Dispatch depends on having a Kubernetes cluster with the Knative components installed (Build, Serving and soon Eventing).  From here build and install dispatch as follows:

1. Set the following environment variables:
    ```bash
    export DOCKER_URL=https://index.docker.io/v1/
    export DOCKER_REPOSITORY="username or repository"
    export DOCKER_USERNAME="username"
    export DOCKER_PASSWORD="password"
    export DISPATCH_NAMESPACE="dispatch-server"
    export RELEASE_NAME="dispatch-server"
    export MINIO_USER="dispatch"
    export MINIO_PASSWORD="dispatch"
    export INGRESS_IP=$(kubectl get service -n istio-system knative-ingressgateway -o json | jq -r .status.loadBalancer.ingress[].ip)
    ```

2. Deploy minio.  Minio is used as a function store:
    ```bash
    helm install --name minio --namespace minio --set accessKey=${MINIO_USER},secretKey=${MINIO_PASSWORD} stable/minio
    ```

3. Build and publish a dispatch image:
    ```bash
    PUSH_IMAGES=1 make images
    ```

4. The previous command will output a configuration file `values.yaml`:
    ```yaml
    image:
      host: username
      tag: v0.1.xx
    minio:
      username: ********
      password: ********
    registry:
      insecure: false
      # Use https://index.docker.io/v1/ for dockerhub
      url: https://index.docker.io/v1/
      repository: repository
      username: ********
      password: ********
    ```

5. Deploy via helm chart (if helm is not installed and initialized, do that first):
    ```bash
    helm upgrade -i --debug ${RELEASE_NAME} ./charts/dispatch --namespace ${DISPATCH_NAMESPACE} -f values.yaml
    ```
    > **NOTE**: Use following to create cluster role binding for tiller:
    >```bash
    >kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
    >```

6. Build the CLI (substitute darwin for linux if needed):
    ```bash
    make cli-darwin
    # Create symlink to binary
    ln -s `pwd`/bin/dispatch-darwin /usr/local/bin/dispatch
    ```

7. Create the Dispatch config:
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
    # point to the config file (could also move to ~/.dispatch/config)
    export DISPATCH_CONFIG=`pwd`/config.json
    ```

8. Test out your install:
    First, create an image:
    ```bash
    dispatch create image python3 dispatchframework/python3-base:0.0.13-knative
    Created function: python3
    ```
    Wait for status READY:
    ```bash
    dispatch get images
       NAME  | DESTINATION | BASEIMAGE | STATUS |         CREATED DATE
    --------------------------------------------------------------------------
     python3 | *********** | ********* | READY  | Tue Sep 25 16:51:35 PDT 2018
    ```
    Create a function:
    ```bash
    dispatch create function --image python3 hello ./examples/python3/hello.py
    Created function: hello
    ```
    Once status is READY:
    ```bash
    dispatch get function
       NAME  | FUNCTIONIMAGE | STATUS |         CREATED DATE
    ----------------------------------------------------------------
      hello  | ************* | READY  | Thu Sep 13 12:41:07 PDT 2018
    ```
    Exec the function:
    ```bash
    dispatch exec hello <<< '{"name": "user"}' | jq .
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
    Create an endpoint:
    ```bash
    dispatch create endpoint get-hello hello --method GET --method POST --path /hello
    ```
    Hit the endpoint with curl:
    ```bash
    curl -v http://${INGRESS_IP}/hello?name=Jon -H 'Host: default.dispatch-server.dispatch.local'
    ```

For a more complete quickstart see the [developer documentation](#documentation)
