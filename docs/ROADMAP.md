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

Currently Dispatch does not contain a user database.  Although there is support for a login action against an IDP (GitHub),
this effectively only ensures that the user has a GitHub account.  Authentication and authorization is a big value
proposition for Dispatch, and the first step is to maintain a database of users.  This is a precursor to full blown
roles and authorization which are also on the roadmap.

An initial implementation should simply support a database and APIs for managing users.  Then the authorization check
should simply ensure that the entity making an API request is included in that user database.  Additionally, the user
metadata should be propogated through the system (associated with the request or event) for auditability.

## 1.2 Roles and Authorization

## 1.3 Applications or Groups

## 1.4 Image Management
