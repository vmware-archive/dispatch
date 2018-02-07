# Dispatch

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

Checkout the detailed [documentation](https://vmware.github.io/dispatch) including a [quickstart guide](https://vmware.github.io/dispatch/guides/quickstart.html).

## Architecture

The diagram below illustrates the different components which make up the Dispatch project:

![initial dispatch architecture diagram](docs/images/dispatch-v1-architecture.png "Initial Architecture")

## Installation

Installing Dispatch is easy once you have a compatible Kubernetes installation.  For instance, to deploy on minikube:

1. Fetch the IP address of minikube as this will be used the host for dispatch services.
```
export DISPATCH_HOST=$(minikube ip)
```

2. Configure the installation:
```
$ cat << EOF > config.yaml
apiGateway:
  host: $DISPATCH_HOST
dispatch:
  host: $DISPATCH_HOST
  debug: true
  skipAuth: true
EOF
```

3. Install Dispatch:
```
$ dispatch install --file config.yaml
```

For a more complete quickstart see the [developer documentation](#documentation)
