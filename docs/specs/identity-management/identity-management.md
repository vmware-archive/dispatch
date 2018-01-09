# Identity and Access Management

## Current Issues

In existing open-source FaaS, one issue is that authentication and authorization are not fully supported. It cannot be
adapted by companies and startups when security/access control of their resources is one of the concerns. Even so, it is
also a pain for them to create separate identities and manage them within the new system.

For example, apache openwhisk don't have fine-grained access control on which user could read/execute which actions. In
most cases, users within a namespace shares the same authentication key and secret. It is not yet integrated with an
identity system.

## Proposed solution

Our Identity and Access Management component is designed to be extensible, plugable, easy-to-use, fine-grained.

The Identity and Access Management is trying to solve two main problems -- Authentication and Authorization.

Based on their specific business need and existing IT infrustructure, the customers should use their best to judgement
to decide on which identity management system they want to integrate with, and which level of access control they want
to enforce.

For **authentication**, we encourage our customer to use their existing identity system. LDAP and OIDC protocols will be
supported, others may on the roadmap. Dispatch will also be shipped along with a built-in identity provider, which
should only be used for local developement/test purpose.

As we use **namespace** to separate different organizations/teams within a single instance of Dispatch, individual
organizations/teams can decide what identity system they want to use within their namespace. We will also have a
``system`` namespace, where resources for the system are placed.

Authentication happens when client request hits our API Gateway or Rest API. The client should specify the identity
provider it want to use, as well as the user's credentials. The gateway will verify the credential with the identity
provider specified. If successfully verified, the request will be forward to the next step, along with the verified user
identity.

**Authorization** or **Access Management** is much more complicated. Resource owner or administrator are required to use
a determinstic way to let Dispatch know _who(principle)_ has the _permission_ to perform what _actions(e.g.
read/write/execute)_ to which _resources(e.g. functions)_.

**Access Management** are resource-based in Dispatch, resources owner (e.g. function owner, service owner) manages
the access control of their resources. In addition, the namespace owner and admins manages identities, and have ultimate
right on access control of resources within the namespace.

## Requirements
- Extensible easy-to-use identity and access management
- Enables the customer to integrate with their own identity management system.
    - use external identity providers instead of a built-in
- Enables the customer to enforce customized access control system.

## User Stories
- As a business/organization owner, I want to take the advantage of the flexibility of serverless, but access to
  resources should be restricted to the people/services that really need it.
- As a developer, I want to have single sign-on experience as what I have with the rest of the services within my
  organization.
- As a system admin, I want to fine-grained control on the restricted resources within Dispatch
- As a end-user, I want to the functions/services easy-to-use but also protected by identity and access controls if
  necessary

## Design
![IAM design](iam.png)

## Milestones/Roadmaps
1. supports OIDC IdP
    - An authentication interface
    - A OIDC plugin
    - Integrated with an external OIDC identity provder (vIDM), with a default access control
    - Identities: only users
    - Access control:
        - only function owner has write permission to the function
        - authenticated user can read/execute any functions within Dispatch
2. supports LDAP IdP
    - A LDAP plugin
    - Be able to integrate with external LDAP identity providers.
3. customized access control:
    - principles: add groups
    - add admin/root permissions

## Summary of Content
1. Terminologies
    - Subject
    - Principal
    - Resources
    - Permissions
2. Authentication Workflow
3. Authorization Workflow

----------------------------------------------------
## 1. Terminologies

### 1.1 Subject
A real person, an external/internal service(function) who _perform actions_ to resources in Dispatch.

Examples:
- Developers
    - related permissions: rwx(read/write/execute) on functions
- Team/Organization admins
    - related permissions: namespace management
- System admins
    - related permissions: namespace management on the namespace 'system'
- End users
    - related permissions: x(execute) on functions
- External/Internal services/functions
    - related permissions: x(execute) on functions

### 1.2 Principle

Digital representations of human users/developers, automation/functions/services. A real person may have mutiple
different principles, based on the different roles/functions they want to perform within Dispatch. E.g. a person who
works as a developer for one function and also acts as an end-user of another functions.

Types:
- User: corresponding to a human users/developers/external services
- Group: associates with a set of permissions, users within the groups inherit the permissions of the group

**Side Note: Namespaces**
(Note: this should be specified on some other design doc)

Namespaces are used to differenciate organizations(public cloud) or teams/business units(private cloud), as different organizations/teams may use different identity systems/protocols. Resources and Identities of different namespace

**namespace 'system'** There's a special namespace called 'system'. It is created when Dispatch is deployed. All
functions are used for life cycle management, and system utilities are put at here. For example, the functions
to create/delete users/groups/namespace, the functions to grant permissions, functions for authentiation/authorizations.

#### (1) Users

A user is corresponding to a real human user/developer or external service/function with respect to Dispatch. For
example, a developer developing a function, a system administrator managing users/groups within an organization/team, a
end-user invoking functions, an external service invoking a weekly function to fetching data in a database.

Depending on capabilities a user has, they should be grant different type of permissions on resources like functions,
namespace and groups.

#### (2) Groups
Groups are a set of users. Permissions can be granted to a group, in which case, all users(group members) in the group
inherits the permission of the group. Groups can NOT be nested. There will be a single level of groups.

### 1.3 Resources

Resources are the objects that, a subject with permission can perform action on. They includes (mainly) functions,
namespaces and principles.

### 1.4 Permissions

Permission specifies **who(principle)** has the _permission_ to perform what **actions(e.g. read/write/execute)** to
which **resources(e.g. functions)**

A _Policy_ is a description file, decribing a list of permissions of users/groups. This can be utilized as a template,
to facilitate the process of create new users, enforce new policies on users/groups.

**Types of permissions:**
- permission on a function
- permission on a group
- permission on a namespace

#### (1) Permissions on a function
- subjects: any principle
- objects: the function
- permissions: read, write, execute, admin, root

Detail:
- any subject is able to create a function
- with read permission, the subject is able to read the function content.
- with write permission, the subject is able to write(update/delete) the function.
- with execute permission, the subject is able to execute the function
- with admin permission, the subject is able to list/add/update/delete any principles' read/write/execute permissions on
  this function
- with root permission, the subject is able to list/add/update/delete any principles' any permissions on this function

Default:
- Any principle, within the namespace, has the read/execute permission on any functions.
- The principle who creates the function has the all permissions on that function.

#### (2) Permissions on a group
- subjects: principles within a group
- objects: the group
- permissions: read/write/root

Details:
- with read permission, the subject is able to list group members
- with write permission, the subject is able to add/remove members to/from the group.
- with root permission, the subject is able to list/add/remove principles' read/write/root permissions of the group

Default:
- any principles within the group, has the read permission on the group
- the principle who creates the group, has the read/write/root permission on the group

#### (3) Permissions on a namespace
- subjects: principles within the namespace
- objects: principles within the namespace
- permissions: read/write/root

Details:
- with read permission, the subject is able to list principles of the namespace
- with write permission, the subject is able to list/add/update/remove principles of the namespace
    - exception: cannot remove principles with root permissions
- with root permission, the subject is able to list/add/update/remove principles' read/write/root permissions of the
  namespace

Default:
- the principle who creates the namespace automatically obtains read/admin/root permissions on the namespace
- any principle within the namespace has the read permission.

---------------------------------------------------------

## 2. Authentication

A client must be authenticated in order to assume its identity within Dispatch.
Authentication is taken place at the API Gateway or Rest API.

### Authentication Methods
- with external Identity Providers
    - External LDAP Server
    - External OpenID Connect(OIDC) Identity Provider
- with Access Key and Secret

Users are authenticated with external Identity Providers or with valid and unexpired access key and access secret.

### Authentication w/ LDAP server

**Workflow**
- client send secure HTTP request with its LDAP username and password
- the request received by a gateway server, forward client credentials to IAM manager.
- IAM manager first verifies if the LDAP server provided is trusted, if not, return with error message
- IAM manager sends _bind_ request with client credentials to the LDAP server
- LDAP server verifies the client credients and responds with user information like name, email, and groups.
- if verfied by LDAP server, IAM manager retreives/creates the corresponding identity from IAM DB, and the identity is
  assumed
- the request is forwarded to the next step along with the assumed identity.

### Authentication w/ OIDC

Prerequisite: Dispatch has already registered as a client application with the OIDC provider, obtains credentials
like ``client id`` and ``client secret``. The OIDC provider should be known and trusted by Dispatch.

**Workflow:**
- client send request to the gateway, and choose authenticate with OIDC
- client is redirected to the authentication endpoint of the OIDC provider it chose
    - along with ``client id`` (as the OIDC provider point of view, we are a client application)
    - and a redirect URI (will be used)
- client is authenticated by the OIDC provider
- client is asked to confirm it authorize us(Dispatch) to use its user data.
- if confirmed, the client will be redirected to the gateway/IAM manager with a ``one time token`` (the redirect URI at
  step 2)
- the IAM manager sends ``client id``, ``client secret``, ``one time token`` to the token endpoint of the OIDC provider,
  exchange for user's ID_token(JWT)
- the IAM manager use the ID_token to look up the corresponding identity in the IAM DB.
- the identity is assumed, forward the request to the next step.

### Authentication w/ Access Key & Secret

Access key and secret pairs are created for external services that want to use functions within Dispatch. They
should managed securely. They can be updated/deleted/suspend by the user. They has expiration time and should be rotated
for best security practice.

**Workflow**
- access key and access secert are created by user when creating a service(principle)
- ...
- client sends secure HTTP request with access key and secret
- gateway received the request, forward client credentials to the IAM manager
- IAM manage verifies the access key and secret with IAM DB
- if verified, the corresponding identity is assumed
- the request is forwarded to the next step along with the assumed identity.

----------------------------------------------------------------------------


## 3. Authorization and Access Control

Dispatch tests the permission the identity the client claimed against the action requested by the client.

When an identity is assumed, (after authenticated with Identity Manager), the identity information is injected into the
the request and the bundle is forward to the event gateway. The access manager within th event gateway verifies if the
identity assumed has the right permission to perform desired action on the requested resource. If that is verified, the
action will be performed.

TODO