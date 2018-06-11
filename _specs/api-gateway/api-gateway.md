---
layout: default
---
# API Gateway & Manager

## Design

### Components

API Manager provides a convenient way to expose function via secure HTTP[S] endpoints.

API Manager is a compontent of Dispatch. Function developers use API manager to create, manage, and expose APIs of their
functions, they are also be able to manage the access control, security requirements their APIs.

API Manager achieves these goals through an API Gateway. Currently, we are using an Open-sourced API Gateway project -
Kong, for more information of Kong, please refer to [Kong Doc](https://getkong.org/docs/)

API Gateway is a service standing in front of Dispatch, it listens to admin request from API manager to configure rules
of how to proxy incoming http requests from end users to upstream functions. The proxy rules are mostly based on http/s
hosts, methods and URIs of the incomming http requests. API Gateway also provides security along with these
functionalities, function developers are able to configure ssl certificates, end-user identity authentiction as well as
access controls. More advanced features, such as rate limiting, request timeout, request size limiting, monitoring and
logging comes later.


### API Manager Specification

#### Basic Functionalities

- The function developers can create/update/delete an API, and link the api to their functions
- The function developers are able to list APIs they have right access to.

APIs will be stored into entity store. API Manager will update any creations/updates/deletions of APIs with the API
Gateway.

APIs and Functions are isolated from each other, functions may have no APIs associated with, and APIs may have no
upstream functions, and any APIs can be pointed to another upstream functions whenever the developers have the right to
do so.

#### Authentiction & Access Control

The function developer should be able to specify the way the end users get authenticated by the API Gateway.

Different ways of authentication:
- public: the API is accessiable anonymously by anyone
- basic: the API is accessible if the end user authenticated by basic username/password
- oidc: the API is accessible if the end user is authenticated by the specified OIDC provider. (note, we probably want
  to require both the developers and end users to use the same OIDC providers)

Access Control:
- API Requests from authenticated end users are honored and responsed.
- API requests from unauthenticated end users are prohibited.


#### HTTPS and certificates:

The function developers are able to specify
- the hostname the function API is served on
- its own ssl certificate and key of their hostname (for HTTPS)
- the function API doesn't support HTTPS connection
- the function API supports both HTTP and HTTPS request from end user
- the function API only supports HTTPS request from end user

However, it is the function developers' responsibility to change their DNS record and point the domain names to the API
Gateway IP address.
