## Dispatch CI configuration

This directory contains configuration of pipelines and jobs executed for CI purposes. Dispatch uses [Concourse CI](https://concourse.ci).
Dispatch Concourse server can be accessed via [ci.dispatchframework.io](https://ci.dispatchframework.io).

### Current pipelines

Dispatch CI consists (as of today) of two pipelines:
* `dispatch-pr` - executed for every Pull Request. runs basic syntax checks, unit tests and coverage.
* `e2e` - executed when PR is merged to master, or for open PRs with `run-e2e` label.

New pipelines will be added as needed.


### Introduction to Concourse

Main way of interacting with Concourse is through its CLI, *fly*. Go to [Downloads](https://concourse.ci/downloads.html) page
to obtain the binary for your platform.

**Note:** ci.dispatchframework.io is read-only, and most of the commands below are accessible to the core team only.

First, you need to login to the server:
```
fly -t dispatch login -c https://ci.dispatchframework.io
```

Then you can see list of pipelines and their jobs:

```
$ fly -t dispatch pipelines
name         paused  public
dispatch-pr  no      yes
```

```
$ fly -t dispatch jobs -p dispatch-pr
name            paused  status     next
dispatch-basic  no      succeeded  n/a 
```

Pipelines in Concourse revolve around three main concepts: *tasks*, *jobs* and *resources* (read more about them [here](http://concourse.ci/concepts.html)).

Tasks are the smallest executable units of work. One of the most useful features of Concourse is ability to trigger a task from command line, including your local changes.

Consider a following definition (taken from ![units/check-syntax.yml](units/check-syntax.yml)) of task:
```yaml
platform: linux

image_resource:
  type: docker-image
  source:
    repository: kars7e/photon-golang-ci
    tag: v0.0.2

inputs:
- name: dispatch

run:
  path: dispatch/ci/units/check-syntax.sh
```

If you run following command (`e` is a shortcut for `execute`):
```
$ fly -t dispatch e -c ci/units/check-syntax.yml --include-ignored -i dispatch=./
executing build 10 at https://ci.dispatchframework.io/builds/10
...
```

the content of local directory will be uploaded as an input named `dispatch`, and the `check-syntax` task will be executed on the CI server.
This way you can check your local changes even before you push your code!

 


