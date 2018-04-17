---
layout: default
---
# Setup Service Account Authentication in Dispatch
Apart from the [Identity Provider (IDP)](setup-authentication.md) workflow, Dispatch also provides service account authentication workflow.

Service account requires valid public/private key pair to authenticate. The private key will be used in client side to sign JWT token and as authentication header in requests. The public key will be used in server side to validate the incoming requests that contain a JWT authentication header.

Use following to generate keys:
```bash
$ openssl genrsa -out <PRIVATE_KEY> 4096
$ openssl rsa -in <PRIVATE_KEY> -pubout -outform PEM -out <PUBLIC_KEY>
```

This document provides instructions for how to setup service account in authentication.

> **NOTE:** Setting up service account authentication requires the privileged access to underlying kubernetes cluster
## 1. Enable Bootstrap Mode with Public Key
There are two ways of enabling bootstrap mode using public key:
### [Option 1] Enable bootstrap mode during Dispatch install
Provide bootstrap user and bootstrap public key in dispatch's install config file. Edit dispatch's install config.yaml to add the information.

```yaml
...
dispatch:
...
  bootstrapUser: <BOOTSTRAP_USER>
  bootstrapPublicKey: <ENCODED_BOOTSTRAP_PUBLIC_KEY>
```
> **NOTE:** ``boostrapPublicKey`` expects a **base64 encoded** public key, failed to provide encoded public key will cause validation failure on server side.

### [Option 2] Enable bootstrap mode using `manage` subcommand
If using `manage` subcommand, ``bootstrapUser`` and ``bootstrapPublicKey`` are not required during dispatch install. After dispatch installation, use following to enable bootstrap mode by providing a bootstrap user name and public key. (Assume ``install.yaml`` is the dispatch install config file)
```bash
$ dispatch manage --enable-bootstrap-mode --bootstrap-user <BOOTSTRAP_USER> --public-key <BOOTSTRAP_PUBLIC_KEY_FILE> -f install.yaml
> bootstrap mode enabled, please turn off in production mode
```
> **NOTEs:**
> * public key will be read from `<BOOTSTRAP_PUBLIC_KEY_FILE>` and encoded by dispatch, no need to encode it manually
> * `manage` operations require k8s cluster access, use `--kubeconfig <K8s_config_file>` flag to speficy, see ``dispatch manage -h`` for more

After enabling bootstrap mode, wait a short period of time (~30 seconds) to sync the information.

## 2. Create Service Account and Policy
Now, Dispatch is in bootstrap mode, we need the `bootstrapPrivateKey`, which is associated with above `bootstrapPublicKey` during install, to authenticate and make requests. Create a public/private key pair for the service account that is going to be created. Use following to create the service account and policy.
```bash
# Create service account - example-user
$ dispatch iam create serviceaccount \
      example-user \
      --public-key ./example-user.key.pub \
      --service-account <BOOTSTRAP_USER> \
      --jwt-private-key <BOOTSTRAP_PRIVATE_KEY_FILE>

# Create policy for the service account
$ dispatch iam create policy \
      example-user-policy \
      --subject example-user --action "*" --resource "*" \
      --service-account <BOOTSTRAP_USER> \
      --jwt-private-key <BOOTSTRAP_PRIVATE_KEY_FILE>
```
where *example-user.key.pub* is the public key file of service account *example-user*, the associated private key will used during make requests using *example-user* service account.

## 3. Disable Bootstrap Mode
Bootstrap mode only allows operations on iam resource, and we have created *example-user* with a policy, now we should disable bootstrap mode by `manage` subcommand:
```bash
$ dispatch manage --disable-bootstrap-mode -f install.yaml
> bootstrap mode disabled
```

## 4. Verify the Service Account
After disabling bootstrap mode, we can verify the service account just created as following, where the *example-user.key* is the associated private key of *example-user.key.pub* that we used to create this service account.
```bash
$ dispatch create -f seed.yaml --service-account example-user --jwt-private-key ../example-user.key
Created BaseImage: nodejs6-base
Created BaseImage: python3-base
Created BaseImage: powershell-base
Created Image: nodejs6
Created Image: python3
Created Image: powershell
Created Function: hello-py
Created Function: http-py
Created Function: hello-js
Created Function: hello-ps1
Created Secret: open-sesame

$ dispatch get base-image --service-account example-user --jwt-private-key ../example-user.key
       NAME       |                   URL                   | STATUS |         CREATED DATE
------------------------------------------------------------------------------------------------
  python3-base    | vmware/dispatch-python3-base:0.0.2-dev1 | READY  | Sat Jan  1 14:40:18 PST 0000
  nodejs6-base    | vmware/dispatch-nodejs6-base:0.0.2-dev1 | READY  | Sat Jan  1 14:40:18 PST 0000
  powershell-base | vmware/dispatch-powershell-base:0.0.3   | READY  | Sat Jan  1 14:40:18 PST 0000
```
