---
---
Dispatch Project Roadmap
========================

### How should I use this document?

This document provides description of items that the project decided to prioritize. This should serve as a reference
point for Dispatch contributors to understand where the project is going, and help determine if a contribution could be
conflicting with some longer term plans.

The fact that a feature isn't listed here doesn't mean that a patch for it will automatically be refused! We are always
happy to receive patches for new cool features we haven't thought about, or didn't judge to be a priority. Please
however understand that such patches might take longer for us to review.

### How can I help?

Short term objectives are listed in [Milestones](https://github.com/vmware/dispatch/milestones) which correspond to a
montly cadence.  Generally speaking any issue which has the label
[Feature](https://github.com/vmware/dispatch/labels/feature) or
[Enhancement](https://github.com/vmware/dispatch/labels/enhancement) are roadmap items. Our goal is to split down the
workload in such way that anybody can jump in and help. Please comment on issues if you want to work on it to avoid
duplicating effort! Similarly, if a maintainer is already assigned on an issue you'd like to participate in, pinging him
on GitHub to offer your help is the best way to go.

### How can I add something to the roadmap?

The roadmap is primarily maintaned by the Dispatch maintainers. We are aiming to be as transparent as possible through
this document and labeling issues. Because roadmap items can have broad effects on the Dispatch project any items added
or changed on this document will be the result of discussions among the maintainers and the author of a proposal.

If you have a proposal which you believe belongs on the roadmap, either raise it in a Issue with the tag "proposal".
This will start a more in depth discussion.

# 1. Features and Refactoring

## 1.1 Users and Authentication

Currently Dispatch does not contain a user database.  Although there is support for a login action against an IDP
(GitHub), this effectively only ensures that the user has a GitHub account.  Authentication and authorization is a big
value proposition for Dispatch, and the first step is to maintain a database of users.  This is a precursor to full
blown roles and authorization which are also on the roadmap.

An initial implementation should simply support a database and APIs for managing users.  Then the authorization check
should simply ensure that the entity making an API request is included in that user database.  Additionally, the user
metadata should be propogated through the system (associated with the request or event) for auditability.

## 1.2 Roles and Authorization

## 1.3 Applications or Groups

Dispatch requires a grouping mechanism in order to better structure resources such as functions and API endpoints.
The suggested name for this grouping is "application", though that is subject to change.  Application is itself a
first class resource, belonging to an organization.  Application should be a required property of the following
resources:

* Function
* API
* Secret
* Subscription (may be implied through function)

Base-images and images may remain tied to an organization.  It's possible that they may optionally be tied to an
application.

Applications should additionally enable per-application hostname (domains?) and certificates on the API gateway.

## 1.4 Image Management

Currently the image support in Dispatch is simply pass-through. See the [image manager spec](image-manager.md) for
feature description.

## 1.5 Multi-Organization Support

Dispatch is a multi-tenant serverless framework.  The top unit of tenancy is "organization".  Because, Dispatch only
supports a single IDP per installation, organization is akin to a business unit with a larger enterprise.  Currently,
a single organization is hard-coded at installation time, however organization is expected to become a configurable
resource.

## 1.6 Scale Dispatch Components

Currently the Dispatch deployment specifies a single instance of each component.  This is fine for development, but
production deployments must scale.  We need to understand how load effects Dispatch and ensure that Dispatch can scale
within reason.  This may require some architectural changes to ensure that work is only done once.

## 1.7 Function Chaining through Events
