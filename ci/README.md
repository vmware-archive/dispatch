# Dispatch CI configuration

This directory contains configuration of pipelines and jobs executed for CI purposes. Dispatch uses [Concourse CI](https://concourse.ci).
Dispatch Concourse server can be accessed via [ci.dispatchframework.io](https://ci.dispatchframework.io).

## Current pipelines

Dispatch CI consists (as of today) of the following pipelines:
* `dispatch-pr` - executed for every Pull Request. runs basic syntax checks, unit tests and coverage.
* `e2e` - executed for open PRs with `run-e2e` label.
* `base-images` - executed for open PRs on every language base images

New pipelines will be added as needed.

### E2E pipeline

E2E pipeline runs 3 jobs:

* `build-images` - compiles binaries, builds docker images and pushes them to a docker registry with tag based on date and commit id.
* `deploy-dispatch` - Deploys dispatch o one of pre-created k8s clusters. Access to these clusters is managed using [Pool resource](https://github.com/concourse/pool-resource).
* `run-tests` - Executes all tests defined in `e2e/tests`. Tests are written using [Bats](https://github.com/sstephenson/bats).

### Base Images pipeline

`base-images` has several jobs:
* `create-gke-cluster` and `delete-gke-cluster`: create and delete gke cluster. Both jobs are triggered manually.
* `install-dispatch`: deploys dispatch running instance, triggered by every Dispatch release and will update the existing dispatch instance automatically.
* `uninstall-dispatch`: uninstalls dispatch, triggered manually.

For each specific language, will have following jobs:
* `build-<LANGUAGE>-base-image`: builds base image based on language pr.
* `test-<LANGUAGE>-base-image`: runs tests on the base image from above pr. And report status back to github pull request.

To add tests for a base image:
* `dispatch/ci/base-images/tests/<LANGUAGE>/task.yml`: concourse ci job yml runs tests. Customize this file to add more language-specific tests preparation.
* `dispatch/ci/base-images/tests/<LANGUAGE>/tests.bats`: bats file contains all tests functions. Tests inside this file will be executed by default.


## Introduction to Concourse

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


### Running task manually
Consider a following definition (taken from ![units/check-syntax.yml](units/check-syntax.yml)) of task:
```yaml
platform: linux

image_resource:
  type: docker-image
  source:
    repository: vmware/dispatch-golang-ci
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

### Hijacking a container
When build fails, it's often tempting to just peek into the container see what happened. To do it with concourse, check pipeline name, job name and build ID, and run:

```
$ fly -t dispatch hijack -j PIPELINE_NAME/JOB_NAME -b BUILD_ID
$ fly -t dispatch hijack -j dispatch-e2e/run-tests -b 11
1: build #11, step: build-cli, type: task
2: build #11, step: dispatch, type: get
3: build #11, step: e2e-tests, type: task
choose a container: 3
root [ /tmp/build/209796a9 ]#
```

It will give you a list of containers that were created for the particular job together with step names they were created for. Just pick a number and voilà - you are in the container.
Containers are retained for some period of time after the build failed and finished its execution.



