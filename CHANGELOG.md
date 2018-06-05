# Changelog
All notable changes to this project will be documented in this file. For more information & examples, check
[What's New](https://vmware.github.io/dispatch/news) section on Dispatch website.

## [Unreleased] - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.15...HEAD)]

### Added
- [BREAKING CHANGE] **Support for contexts when working with Dispatch CLI.**
Dispatch config files now have contexts, allowing user to work with multiple instances of Dispatch at the same time.
This change requires config file to be context-aware (old version of config won't work with new CLI). [PR #478](https://github.com/vmware/dispatch/pull/478).
- **Initial support for function timeout.** 
When creating a function, user can now specify a timeout (in milliseconds).
When executing a function, it will be terminated if takes longer than the specified timeout. Note that this is language-dependent (language pack must support timeout).
[PR #487](https://github.com/vmware/dispatch/pull/487).

### Fixed
- [[Issue #472](https://github.com/vmware/dispatch/issues/472)] **Send empty response in API gateway when function does not return anything.**
Previously, API gateway would return a complete output of function execution from function manager if function itself didn't return any content.
This exposed a lot of arguably private information about the internal implementation of the function.
From now on, the API gateway will return an empty response if function output is null. [PR #477](https://github.com/vmware/dispatch/pull/477).


## [0.1.15] - 2017-05-24 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.14...v0.1.15)] [[What's new](https://vmware.github.io/dispatch/2018/05/23/v0-1-15-release.html)]

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


## [0.1.14] - 2017-05-15 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.13...v0.1.14)]

### Added
- **Support for private image registries that require auth for pulling (OpenFaaS only).**
Dispatch can now properly when using external private registry that requires credentials when pulling an image.
This feature is only available when using OpenFaaS backend. [PR #438](https://github.com/vmware/dispatch/pull/438).

### Fixed
- **Improved IAM documentation.**
Bootstrap mode and service account management should now be clearer. [PR #425](https://github.com/vmware/dispatch/pull/425).