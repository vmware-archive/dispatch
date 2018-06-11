---
layout: default
---
# Dispatch Secret Store
A spec for how to store secrets necessary for Functions

## Current Issues
FaaS / Serverless implementations require access to external resources and services to perform non trivial tasks. All
but the most public read-only resources and services will require their own authentication and authorization. The
requirement for authentication with external resources raises the question of securely managing authentication secrets
used by functions. We propose a secure secrets store to answer this question.

## Proposed Solution
The secret store will provide a simple API to associate secrets with functions, users, organizations and/or namespaces.
This API will abstract away the concrete implementation of the backing storage and allow for pluggable secret injection
based on the underlying implementation. For example, the API can access and store secrets from a store such as
Kubernetes Secrets or Hashicorp's Vault and inject those secrets securely into the function container wether it is
running on OpenFaaS or Riff or some other implementation.

As a guiding principle, functions should not include secrets directly in their definitions. The FaaS implementation
should provide the secret at function invocation. Providing secrets in this way has several benefits. It removes
redundant declaration of secrets across functions, it eliminates the threat posed by storing secrets in version control
and it allows for more centralized management of secrets.

A user can access these secrets within his/her function definition in a couple of ways. A user can access secrets via
environment variables set in the container running the function, or they can access them as extra parameters to the
function itself.

## User Stories
- A function should have its required secrets injected at function invocation
- A function should not have access to secrets not associated with the function. Further more functions should not have access to any secret not directly associated with it even if the secret belongs to the same namespace, organization, etc.
- A user / administrator can input secrets without needing to know the underlying storage mechanism.
- A function should not require changes if the implementation is changed.

## Summary of Content
1. Terminology
2. Dependencies
3. Challenges
4. Workflows

---
## 1. Terminology

### Secret
Any key/value pair used to map sensitive information. For example,
- Username/Password
  - username : white_rabbit
  - password : 1_am_l8
- API Token
  - api_token : 98E2AB51A64298712BCFC531F4E81

See Identity and Access Management for definitions of namespace, user, group, and permission

## 2. Dependencies
The implementation of the secret store depends heavily on the implementation of the user authentication/authorization in
the product generally. Secrets require permissions similar to those of functions but with extra permissions defining
which functions have access to which secrets.

## 3. Challenges
Protecting sensitive data presents many challenges. The secret store must limit the surface area for attacks trying to
steal sensitive data. The nature of distributed microservice architecture requires more protection than traditional
monolithic services.

Secrets stored in a centrally managed store must be transmitted to worker nodes performing function execution. The
channel used to transfer these secrets must be encrypted. The secrets should not be accessible on nodes that do not
require the secret. This includes purge secrets from nodes when the function requiring them is removed from the node.
