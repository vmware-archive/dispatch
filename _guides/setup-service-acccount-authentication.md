---
layout: default
---
# Service Accounts in Dispatch

Users in Dispatch are managed by an external [Identity Provider (IDP)]({{ '/documentation/guides/setup-authentication' | relative_url }}) e.g OpenID Connect provider like vIDM, Dex etc. As such, when users login to dispatch they will be redirected to the configured OIDC provider for authentication.
This authentication step involves the end-user and that works in their best interest. For non-human users, like a CI/CD system, a third-party application or service interacting with Dispatch API's, a special kind of user account is required.
Dispatch calls them as Service Accounts and they are completely managed by Dispatch's Identity Manager.

Service accounts can be created and managed using Dispatch CLI or API's. Service account authentication involves the generation of a JWT bearer token by the client and the token must be signed using one of the "RS256/384/512" - JSON Web Signature algorithms. If you are using the CLI, most of the token generation and signing is already taken care.

## 1. Generate a RSA key pair
Before creating a service account, we need to generate a RSA public/private key pair. You can use any key length greater than 2048 bits.  The private key will be used by the client to sign the JWT token and specify it as a bearer token in the Authorization HTTP header of any API request. The public key will be used on server side to validate the JWT token in the API requests.
Hence, Dispatch only requires you to specify the public key when creating a service account in Dispatch. Make sure to keep the private key safe with the process or application that will interact with Dispatch.

Use following openssl commands to generate a key pair:

```bash
$ openssl genrsa -out <PRIVATE_KEY> 4096
$ openssl rsa -in <PRIVATE_KEY> -pubout -outform PEM -out <PUBLIC_KEY>
```

## 2. Create Service Account
Login to dispatch and use the public key from the previous setup to create a service account:

```bash
# Create service account - example-svc-account
$ dispatch iam create serviceaccount \
      example-svc-account \
      --public-key ./example-user.key.pub
> Create service account: example-svc-account
```
The name of the service account is important when generating the jwt token and it must be specified in the issuer field `iss` of the jwt payload. More about this is covered in the usage section below.

## 3. Create a Policy
Similar to user accounts, service accounts are not useful unless an access policy is created. Create a policy and associate it with the service account:
```bash
# Create policy for the service account
$ dispatch iam create policy \
      example-svc-policy \
      --subject example-svc-account --action "*" --resource "function" \
      --service-account <SERVICE_ACCOUNT_NAME>
> Created policy: example-svc-policy
```

## 4. Using the Service Account
When invoking dispatch commands, specify the private key associated with the *example-user.key.pub* that we used to create this service account in the `--jwt-private-key` flag. Dispatch CLI will
use this private key to sign the generated JWT token.

```bash
$ dispatch create -f seed.yaml --service-account example-svc-account --jwt-private-key ../example-user.key
Created BaseImage: nodejs-base
Created BaseImage: python3-base
Created BaseImage: powershell-base
Created Image: nodejs
Created Image: python3
Created Image: powershell
Created Function: hello-py
Created Function: http-py
Created Function: hello-js
Created Function: hello-ps1
Created Secret: open-sesame

$ dispatch get base-image --service-account example-svc-account --jwt-private-key ../example-user.key
       NAME       |                   URL                   | STATUS |         CREATED DATE
------------------------------------------------------------------------------------------------
  python3-base    | dispatchframework/python3-base:0.0.7    | READY  | Sat Jan  1 14:40:18 PST 0000
  nodejs-base     | dispatchframework/nodejs-base:0.0.6     | READY  | Sat Jan  1 14:40:18 PST 0000
  powershell-base | dispatchframework/powershell-base:0.0.8 | READY  | Sat Jan  1 14:40:18 PST 0000
```

If you are directly calling the API instead of using the CLI, you need to create a JWT payload as follows:

```json
{
 "iss": "example-svc-account",
 "iat": "1525930134",
 "exp": "1525865330"
}
```
, then sign the payload with the associated private key using one of RS256/384/512 algorithms and present it in the HTTP Authorization header as a bearer token e.g `Authorization : Bearer <JWT_TOKEN>`. You can learn more about JWT tokens [here](https://jwt.io/introduction/).
