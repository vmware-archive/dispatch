# Changelog
All notable changes to this project will be documented in this file. For more information & examples, check
[What's New](https://vmware.github.io/dispatch/news) section on Dispatch website.

## [Unreleased] - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.22...HEAD)]

### Added

### Fixed

### Removed

## [0.1.22] - 2018-07-18 [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.21...v0.1.22)]

### Fixed
- **Fix dispatch uninstall cmd ignoring config file** `dispatch uninstall` cmd was ignoring the namespace definition in the config file. This patch fixes it and adds some unit test as the command is broken more often than not.
- **Bump version of PhotonOS in our images** An ancient version of PhotonOS image was used. Updated to `vmware/photon2:20180620`.

### Removed
- **Move source code out of function entity** Function entity used to hold the .tar.gz archive of the source code folder (or file) from which the function was built.

## [0.1.21] - 2018-07-10 [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.20...v0.1.21)]

### Added

- [[Issue #518](https://github.com/vmware/dispatch/issues/518)] **Single-binary, local version of Dispatch Server:**
This release includes a single-binary dispatch-server. You can run this server locally on your desktop without a need
to provision Kubernetes - the only requirement is Docker. This should cover use cases like local development, proofs of concept,
or a small deployment for personal needs. To use it, simply download the `dispatch-server` binary for your platform,
and run `dispatch-server local`.
    
    *Note:* The local version supports all commands/resources except:
    - event drivers
    - services 
- **Add Org to bulk create CLI** Orgs can now be defined in yaml files and populated in Dispatch using `dispatch create -f seed.yaml`
- **Ingress class option** Use this option to target a different ingress controller

### Fixed

- **Fixed entity store bug** Fixed a bug where the libkv-based entity store would return entities of different types if they shared the same prefix
- **Fixed chart issue** Fixed an issue where the resources specified for an individual sub-chart didn't take precedence over the global resources

### Removed

- **Removed internal vCenter event driver** The vCenter event driver is no longer built in to Dispatch. You can access the vCenter event driver here https://github.com/dispatchframework/dispatch-events-vcenter

## [0.1.20] - 2018-07-03 - [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.19...v0.1.20)]

### Added

- [[Issue #465](https://github.com/vmware/dispatch/issues/465)] **Tail/follow function runs as they happen.** When
retrieving runs you can now specify the `--last` flag to get the latest executed run. You can also now specify the
`--follow` or `-f` flag to follow/tail runs as they occur. [PR #537](https://github.com/vmware/dispatch/pull/537)

- **API Gateway endpoints now namespaced with the org name** Creates the user-specified API paths under the org name.
This is because we share an API Gateway and the paths from different org's can overlap. For production uses, user
usually specifies a hostname for the API's.

### Fixed

## [0.1.19] - 2018-06-26 [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.18...v0.1.19)] [[What's new](https://vmware.github.io/dispatch/2018/06/26/v0-1-19-release.html)]

### Added

- **Enable Exposed Event Drivers** Events may now be pushed (as opposed to pulled) from event sources to Dispatch.  An *exposed*
driver will create an ingress route and service to recieve events.

## [0.1.18] - 2018-06-19 [[Git compare](https://github.com/vmware/dispatch/compare/v0.1.17...v0.1.18)] [[What's new](https://vmware.github.io/dispatch/2018/06/19/v0-1-18-release.html)]

### Fixed

- **Limited Multi-Tenancy Support.** Dispatch now supports multi-tenancy with the ability to create multiple organizations and define policies to gives access to users on those organizations.
This is a limited functionality and does not isolate the function execution environments within the underlying FaaS engine or define network policies for the function pods running on Kubernetes.[PR #510](https://github.com/vmware/dispatch/pull/510) [PR #529](https://github.com/vmware/dispatch/pull/529).
- **Fix single file Java support.** Previously Java would fail for single files if no handler argument is provided. [PR #517](https://github.com/vmware/dispatch/pull/517)
- [[Issue #483](https://github.com/vmware/dispatch/issues/483)] **No helpful error output when an error occurs during function create.**
Added a `reason` field to the function object to hold function creation errors. [PR #513](https://github.com/vmware/dispatch/pull/513)

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