# Changelog
All notable changes to this project will be documented in this file. For more information & examples, check
[What's New](https://vmware.github.io/dispatch/news) section on Dispatch website.


## [Unreleased] - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.17...HEAD)]

### Added

### Fixed

## [0.1.17] - 2018-06-12 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.16...v0.1.17)] [[What's new](https://vmware.github.io/dispatch/2018/06/12/v0-1-17-release.html)]

### Added

- **Improved IAM bootstrap workflow.** New CLI Command `dispatch manage bootstrap` to automatically bootstrap Dispatch with a new organization, service account and policies upon installation. This replaces
the `dispatch manage --enable-bootstrap-mode` command that required the user to manually enter the bootstrap mode to create the authorization accounts and policies. [PR #501](https://github.com/vmware/dispatch/pull/501).
- **New CLI Command to print the versions.** A New CLI Command `dispatch version` to print the client and server versions has been introduced. As part of this change a new `/v1/version` API has been added
to the Identity Manager service that returns the current server version. The client version is embedded into the `dispatch` binary. [PR #500](https://github.com/vmware/dispatch/pull/500).
- **New CLI command to create seed images.** A New CLI Command `dispatch create seed-images` to create base-images and images of commonly used languages in the dispatch community. The revision of the created base-images/images
corresponds to that of the CLI's. [PR #507](https://github.com/vmware/dispatch/pull/507)
 
### Fixed

- [[Issue #251](https://github.com/vmware/dispatch/issues/251)] **Deleting image deletes its base-image if they have the same name.**
Previously when deleting an image with the same name as a base-image, the base-image would also be deleted at the same time. It's fixed now. [PR #504](https://github.com/vmware/dispatch/pull/504).

## [0.1.16] - 2018-06-06 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.15...v0.1.16)] [[What's new](https://vmware.github.io/dispatch/2018/06/06/v0-1-16-release.html)]

### Added
- [BREAKING CHANGE] **Support for contexts when working with Dispatch CLI.**
Dispatch config files now have contexts, allowing user to work with multiple instances of Dispatch at the same time.
This change requires config file to be context-aware (old version of config won't work with new CLI). [PR #478](https://github.com/vmware/dispatch/pull/478).
- **Initial support for function timeout.** 
When creating a function, user can now specify a timeout (in milliseconds).
When executing a function, it will be terminated if takes longer than the specified timeout. Note that this is language-dependent (language pack must support timeout).
[PR #487](https://github.com/vmware/dispatch/pull/487).
- **CLI login now works with service accounts.**
Previously, `dispatch login` command only worked with OIDC users (e.g. when authentication has been configured using GitHub auth provider).
Now you can run `dispatch login --service-account test-svc-account --jwt-private-key <path-to-key-file>` and the auth data will be stored in the dispatch config file.
there is also a corresponding `dispatch logout` command to clear the credentials. [PR #460](https://github.com/vmware/dispatch/pull/460).
- **Batch resource creation now supports all resource types.**
One of the handy features of Dispatch CLI is batch creation of resources, where user can specify multiple resources using YAML file, and then create them all
using `dispatch create -f file.yaml` command. In this release, the batch file supports all resource types 
(previously event drivers, driver types, event subscriptions and APIs were missing). [PR #495](https://github.com/vmware/dispatch/pull/495/files).

### Fixed
- [[Issue #472](https://github.com/vmware/dispatch/issues/472)] **Send empty response in API gateway when function does not return anything.**
Previously, API gateway would return a complete output of function execution from function manager if function itself didn't return any content.
This exposed a lot of arguably private information about the internal implementation of the function.
From now on, the API gateway will return an empty response if function output is null. [PR #477](https://github.com/vmware/dispatch/pull/477).
- [[Issue #486](https://github.com/vmware/dispatch/issues/486)] **Event Driver creation failed when secrets were used.**
Previous release introduced a regression in event driver creation if driver was configured with secrets. It's now fixed.
[PR #488](https://github.com/vmware/dispatch/pull/488).
- [[Issue #485](https://github.com/vmware/dispatch/issues/485)] **Dispatch resources displayed incorrect timestamps.**
The timestamps for Dispatch resources (functions, images, etc.) displayed incorrect dates, i.e. dates that did not correspond to the actual
dates of creation/modification of resources. In this release, the dates should now be properly saved and displayed. 
NOTE: due the nature of this bug, only new deployments will notice the fix. [PR #494](https://github.com/vmware/dispatch/pull/494).
 
## [0.1.15] - 2018-05-24 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.14...v0.1.15)] [[What's new](https://vmware.github.io/dispatch/2018/05/23/v0-1-15-release.html)]

### Added
- **New function manager backend using [Kubeless](https://github.com/kubeless/kubeless)**.
You can now run your functions in Dispatch using Kubeless (many thanks to [Andres](https://github.com/andresmgot) for the contribution).
Set the `faas` option to `kubeless` in your install config to deploy Dispatch with kubeless. More examples with Kubeless coming soon. 
- **Support for providing entire directory instead of single file when creating a function**. 
You can now specify a directory instead of single file when creating a function. 
See [What's new in v0.1.15] for more details on using this feature. PRs [#470](https://github.com/vmware/dispatch/pull/470), [#474](https://github.com/vmware/dispatch/pull/474).
- **New CLI command `dispatch log` to print service logs**.
Previously, to check logs of any of dispatch services, one had to find the name of the Kubernetes Pod corresponding to the service,
and use kubectl to grab the logs.
With this change, getting logs for function manager is as simple as running `dispatch log function-manager` (similar for other services).
It supports the follow mode (`-f`) as well! [PR #445](https://github.com/vmware/dispatch/pull/445).
- **Support for HTTPS via Let's Encrypt and Cert Manager**.
API gateway can now be exposed using a hostname with a valid, signed certificate, courtesy of Let's Encrypt.
For details, see [the PR #427](https://github.com/vmware/dispatch/pull/427). 
- **Improved support for Tracing.** Dispatch now propagates tracing context within services, resulting in better view of code involved in handling the request. [PR #461](https://github.com/vmware/dispatch/pull/461).

### Changed
- **Improved handling of function execution errors.** 
Functions can now return different errors when execution fails, depending on when and why the error happened.
See [Error handling specification](https://github.com/vmware/dispatch/blob/5f1043a55018fbadbbc5e1fbf507a5f2a9fc9121/docs/_specs/error-handling/error-handling.md) for more details.
PRs [#423](https://github.com/vmware/dispatch/pull/423), [#424](https://github.com/vmware/dispatch/pull/424).

### Fixed
- **`Gopkg.toml` is configured to prune unused dependency files.**
Previously, after running `dep ensure`, one had to use separate [`prune`](https://github.com/imikushin/prune) command to clean the `vendor` directory.
Now the prunning is done as a part of `dep ensure`. This change should make it easier for contributors to add new dependencies.
[PR #441](https://github.com/vmware/dispatch/pull/441).
- **Dispatch services now use flattened swagger spec.**
Previously, spec had to be flattened as an extra step during spec generation. This change should make the `make generate` step
much shorter. [PR #443](https://github.com/vmware/dispatch/pull/443). 
- **Dispatch built-in vcenter event driver now properly reads environment variables.** [PR #456](https://github.com/vmware/dispatch/pull/456).


## [0.1.14] - 2018-05-15 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.13...v0.1.14)]

### Added
- **Support for private image registries that require auth for pulling (OpenFaaS only).**
Dispatch can now properly when using external private registry that requires credentials when pulling an image.
This feature is only available when using OpenFaaS backend. [PR #438](https://github.com/vmware/dispatch/pull/438).

### Fixed
- **Improved IAM documentation.**
Bootstrap mode and service account management should now be clearer. [PR #425](https://github.com/vmware/dispatch/pull/425).