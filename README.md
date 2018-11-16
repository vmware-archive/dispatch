![Dispatch](docs/assets/images/logo-large.png "Dispatch Logo")

> **ATTENTION**: This Readme is for Dispatch Solo version, for Dispatch Knative Version, please go to [Master](https://github.com/vmware/dispatch) branch.

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

**Dispatch-Solo** is a packaged up Dispatch distribution.  The intent is to lower the barrier for users to try out Dispatch and get feedback.

> **ATTENTION**: Dispatch-Solo is not intended as a production service and Dispatch-Solo is not upgradable to Dispatch-Knative

## Documentation

Checkout the detailed [documentation](https://vmware.github.io/dispatch) including a [quickstart guide](https://vmware.github.io/dispatch/documentation/guides/quickstart).

## Architecture

The diagram below illustrates the different components which make up the Dispatch project:

![solo dispatch architecture diagram](docs/assets/images/solo-arch.png "Initial Architecture")

## Installation
Please go to [Dispatch-Solo-OVA](https://github.com/vmware/dispatch/wiki/Dispatch-Solo-OVA) Wiki page.

## For Developer

### Requirement
1. Golang installed.
2. Docker installed.
3. Clone source code to $GOPATH: you can clone from `https://github.com/vmware/dispatch.git`

### Build dispatch binaries

#### For Mac OS
```bash
make darwin
cp ./bin/dispatch-darwin /usr/local/bin/dispatch
# Start dispatch Server
./bin/dispatch-server-darwin
```

#### For Linux
```bash
make linux
cp ./bin/dispatch-linux /usr/local/bin/dispatch
# Start dispatch Server
./bin/dispatch-server-linux
```
### Run E2E test
Please go to [Running E2E Tests](https://github.com/vmware/dispatch/wiki/Running-E2E-Tests) Wiki Page

### Contributing

You are invited to contribute new features, fixes, or updates, large or small; we are always thrilled to receive pull
requests, and do our best to process them as fast as we can. If you wish to contribute code and you have not signed our
contributor license agreement (CLA), our bot will update the issue when you open a [Pull
Request](https://help.github.com/articles/creating-a-pull-request). For any questions about the CLA process, please
refer to our [FAQ](https://cla.vmware.com/faq).

Before you start to code, we recommend discussing your plans through a  [GitHub
issue](https://github.com/vmware/dispatch/issues) or discuss it first with the official project
[maintainers](AUTHORS.md) via the [#Dispatch Slack Channel](https://vmwarecode.slack.com/messages/dispatch/), especially
for more ambitious contributions. This gives other contributors a chance to point you in the right direction, give you
feedback on your design, and help you find out if someone else is working on the same thing.

## License

Dispatch is available under the [Apache 2 license](LICENSE).