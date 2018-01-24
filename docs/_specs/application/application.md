---
layout: default
---

# Application-scoped Dispatch

## Introduction

Within Dispatch, we believe the concept of “Application” is a very important aspect. In a non-trivial deployment of dispatch, Applications group and isolate resources of different function unit, provide the ability of searching and filtering resources; they also facilitate permission management and access controls.

In addition to those basic functionalities, the concept of Application also provides a step forwards to advanced features of cloud native apps, e.g. lifecycle management (i.e. creation/delete, import and export Applications, scaling up and down based on demand/usage);, function execution workflow management, visualized workflow GUI; application level monitoring and analyzing; function/application external dependencies management.

In this design documents, we outline several design concerns and implementation decisions with regards to the concept of a Serverless Application.

## Manage Entities with Application

In this section, we discuss the relationships of Application and Entities, how Entities will be organized and managed by Applications.


### Relationship of Applications and Entities

We identified three kinds of relationships of Application and Entities. First, many-to-many, applications consists of multiple entities and one entity can be used by multiple Applications. Second, one-to-many, an entity is only used by one Application; lastly, some entities have no obvious relationship with Application, let's call them system-scoped entities. They may be used by applications or entities of an application.

We will discuss all kinds of Entities that Dispatch current manages case by case.

### Base Images

The intention of Base Images is to provide an easy way for dispatch system admin to easily manage lifecycle of function runtimes (operation systems and languages runtime) in a centralized manner and minimize developers’ effects of managing infrastructures. Developers of different Applications are encouraged to leverage existing Base Images in the dispatch system to build their own functions. Therefore, Base Images should be shared by as many applications as they can, and they should be system scoped entities.

### Identities & Access Controls - Users & Groups

It hasn’t be decided whether or not we will manage users, groups and access controls within dispatch or with the help of an existing framework. If we decide we should include them as dispatch managed entities, they should be system-scoped. In addition to it, there should be some description/information on which users/groups have permissions on which applications.

### Functions, Images & Secrets

In our vision of Dispatch, functions, images and secrets are more like packages, libraries or source code of a traditional application. They are not the runtime resources of a single Application instance. With this point of perspective, we think there are potential use cases that a function or function image could be used by multiple applications. However, note there indeed are some debates going on here. Rather than categorizing them deterministically into one single basket, we decide to keep it flexible in implementation, and iterate quickly as we see sufficient use cases in the future.

Currently, we will temporarily put them into the one-to-many category. However, again, we will be flexible and extensible in implementation detail.


### Function Runs

Function Runs are entities generated from function executations. They are runtime entities and they should belongs to a single Application.

### APIs

APIs are considered tightly bound with an Application. Different applications have no reason to share one API. Therefore, one API should stay in the same Application as its bounded Function.

### Event Drivers, Topics & Subscriptions

There are indeed some use cases that some events maybe shared by multiple applications, (e.g. events from the dispatch or kubernetes internal, VM lifecycle events may be in demand for several applications).

However, in some other use cases, events maybe only be used by specific applications, e.g. a PR build failure event triggers a report to a internal Slack channel, as we can image, it will almost be used by only one dispatch application.

Both shared one-to-many and many-to-many events are valid use cases. When there’s not enough real use cases it is hard to decide which is dominant and which is rare. Again, we decide to keep it flexible at this point. In the first iteration, we will set them as private to an Application for now.

As a side note, although Event Topics are not a dispatch entity now, we should also consider to scope them by Application (especially if we decide that Event Subscriptions should be scoped). For example, add the Application name as a prefix of a topic name.

## Implementation Detail


### The Application Entity

As most entities are organized by applications, an Application itself also has a life cycle. To manage applications, it is best to manage and maintain it with the existing Entity Store.

The Application Entity will be private to a single Application. With respect to management, they will only be managed by system admin and the users who has right permissions on it. Again, as access control is not the main concern of the this article, we won’t go into detail about that.

An Application Entity should have references to a set of Entities it manages/uses. It facilitates further expansion of advanced features with regards to a serverless application. In the next step, there might be attributes related to application-level access control, resource limits, as well as schema/definition of the workflow.

### The Application Tags in other Entities

We are going to associate an Entity to Applications with tags. The advantage of using tags instead of an fixed struct field is that it provides a flexible way to extend the relationship of Applications and Entities.

### Operations related to Applications

Here we list a set of operations required, in additional to existing ones. They will be used to manage Applications as well as filter entities by Applications.

- list applications
    - e.g. ``dispatch get applications``
- manage an application
    - create, delete and update
- list entities of an applications
    - e.g. ``dispatch get --application example-app``
- get/list a specific entity of an applications
    - e.g. ``dispatch get functions --application example-app``
- get entity by name and application
    - e.g. ``dispatch get function hello --application example-app``
- list all entities of all applications
    - e.g. ``dispatch get --all-applications``
