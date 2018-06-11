---
layout: default
---
# Setup Dispatch with Authentication & Authorization

If you previously followed the [Quickstart](quickstart.md) guide, it setups a Dispatch installation without any user authentication
or authorization. This guide will help you setup Dispatch with an Identity Provider and additionally configure
authorization policies.

Dispatch is a serverless framework that requires developers and end-users to authenticate themselves before they are
using the platform.

Users in Dispatch are managed by an external Identity Provider (IDP) like Github, Google Identity Platform or other enterprise [OpenID Connect](https://openid.net/connect/) (OIDC) providers like vIDM.
OpenID Connect enhances OAuth 2.0 authorization protocol workflow to support authentication.
When users login to dispatch they will be redirected to the configured OIDC provider for authentication.

This document provides instructions for how to integrate specific IDP's into Dispatch and setup policies.

## 1. Setup Dispatch install config without `skipAuth`

If you followed the  [Quickstart](quickstart.md) guide, you would have setup Dispatch by skipping authentication i.e with `skipAuth: true` in install config.yaml file. You need to
ensure `skipAuth` is unset or set to `false` in the install config.yaml file before proceeding with installing Dispatch.

Your config.yaml file may now look similar to 

```yaml
dispatch:
  apiGateway:
    host: <DISPATCH_HOST>
  dispatch:
    host: <DISPATCH_HOST>
    debug: true
```

## 2. Create An OAuth Client App with your Identity Provider

### Using Github

Login to your Github Account and go to [Github developer portal](https://github.com/settings/developers)

Click ``New OAuth App`` Button to create a new client app.

- Application name: for your reference only, e.g. ``dev-dispatch-app``
- Homepage URL: the hostname or IP address of your Dispatch deployment, e.g. ``https://dev.dispatch.vmware.com``
- Authorization callback URL: <homepage-url>/v1/iam/oauth2/callback, e.g. ``https://dev.dispatch.vmware.com/v1/iam/oauth2/callback``

Click ``Register application`` and now the client app is created

You should see ``Client ID`` and ``Client Secret`` in the next page, they are the credentials you will use in the next
step.

Edit dispatch's install config.yaml to add the information of the Identity Provider to the `oauth2proxy` key.


```yaml
dispatch:
  apiGateway:
    host: <DISPATCH_HOST>
  dispatch:
    host: <DISPATCH_HOST>
    debug: true
  oauth2Proxy:
    provider: github
    clientID: <client-id>
    clientSecret: <client-secret>
```

### Using Google Identity Platform

Login to your [Google API Console](https://console.developers.google.com/) and [Create a Project](https://console.developers.google.com/projectcreate) if you don't have one already. You must setup a project in order to proceed to the next steps.

* Navigate to the _API and Services_ -> _[Credentials](https://console.developers.google.com/apis/credentials)_ page from the left menu of your project's home page.
* Enter a Product name e.g ``dispatch-dev-app`` in the OAuth2 Consent Screen tab
* Click on ``Create credentials > OAuth client ID``
* Choose ``Application type`` as ``Web Application``
* Provide a name to the client app
* Specify the Authorization Redirect URI as ``https://<dispatch_host>/v1/iam/oauth2/callback`` where ``dispatch_host`` is the hostname of IP address of your dispatch deployment. e.g ``https://dev.dispatch.vmware.com/v1/iam/oauth2/callback``
  - If Dispatch is running on a non-standard port (i.e. anything other than 443), the Authoriziation Redirect URI must include that port.  You will need
    to update the Redirect URI after Dispatch is installed to get that port.
* Click ``Create``

You should see ``Client ID`` and ``Client Secret`` in the next page, they are the credentials you will use in the next
step.

For more detailed information visit Google's [Setting up an OAuth2 App page](https://developers.google.com/identity/protocols/OpenIDConnect#appsetup)

Edit dispatch's install config.yaml to add the information of the Identity Provider to the `oauth2proxy` key.

```yaml
dispatch:
  apiGateway:
    host: <DISPATCH_HOST>
  dispatch:
    host: <DISPATCH_HOST>
    debug: true
  oauth2Proxy:
    provider: oidc
    oidcIssuerURL: https://accounts.google.com
    clientID: XYZ.apps.googleusercontent.com
    clientSecret: <client-secret>
```

###  Other OpenID Connect Providers

Dispatch supports OpenID Connect compliant Identity Providers for providing authentication.

The steps to create an ``OAuth2 Client App`` varies by the provider and hence please refer to the provider documentation.
Most likely you just need the _Authorization Redirect URI_ which is ``https://<dispatch_host>/v1/iam/oauth2/callback`` where ``dispatch_host`` is the hostname of IP address of your dispatch deployment.

Once you have secured the ``Client ID`` and ``Client Secret`` from your provider,
edit dispatch's install `config.yaml` to add the information of the Identity Provider. You also need the `Issuer URL` of your ODIC compliant Identity provider.

```yaml
dispatch:
  apiGateway:
    host: <DISPATCH_HOST>
  dispatch:
    host: <DISPATCH_HOST>
    debug: true
  oauth2Proxy:
    provider: oidc
    oidcIssuerURL: <OIDC Issuer URL>
    clientID: <client-id>
    clientSecret: <client-secret>
```


## 3. Create Cookie Secret (Optional)

Dispatch uses HTTP session cookies to keep track of users. It is optional to encrypt the cookie sent to the end users, but it is highly recommended for security reasons.

To generate a random secret key:
```bash
$ python -c 'import os,base64; print base64.b64encode(os.urandom(16))'
YVBLBQXd4CZo1vnUTSM/3w==
```

Specify the cookie secret in the install config.yaml's `oauth2proxy` section
```yaml
dispatch:
  apiGateway:
    host: <DISPATCH_HOST>
  dispatch:
    host: <DISPATCH_HOST>
    debug: true
  oauth2Proxy:
    ...
    cookieSecret: YVBLBQXd4CZo1vnUTSM/3w==
```

## 4. Install Dispatch

Install Dispatch with

```bash
dispatch install -f config.yaml
```

## 5. Bootstrap Dispatch IAM

After Dispatch is installed successfully, you need to bootstrap it's Identity Manager with some initial authorization policies. This is akin to setting up your new laptop with an administrative account.
If you try to `dispatch login` without any authorization policies in place, even if the authentication is successful with the configured Identity Provider (e.g github), users will be denied access to protected resources in dispatch.

> **NOTE:** You still need to have access to the Kubernetes cluster on which dispatch was installed since bootstrap is a privileged operation.
 
The goal of bootstrap is to setup the initial authorization policies for a specified user such that the user can then use the normal dispatch commands to setup additional policies. 

In order to proceed, you need to identify the email address of the user account from your Identity Provider that will be used to setup the initial authorization policies e.g. with GitHub, this is the primary email address associated with your github account.
With OpenID Connect providers, this is normally the email address associated with your user profile.

Run the bootstrap command with
```bash
dispatch manage bootstrap --bootstrap-user <xyz@example.com>
```

The bootstrap command forces the system to enter a special mode that bypasses normal authentication and allows us to setup initial policies. The command then disables that mode.

> **NOTE:** Please ensure to see the bootstrap mode is disabled as it can leave your installation vulnerable.
>
> If the command fails to disable bootstrap mode, you can manually issue the following command to disable it.
> ```bash
> dispatch manage bootstrap --disable
> ```

## 6. Login to Dispatch

Login to dispatch with
```bash
dispatch login
```

You will now be redirected to your configured Identity Provider for authentication on a browser.

Sign-in to your Identity Provider as the `--bootrstrap-user` that you configured in the previous step. Upon successful authentication, you should see the following response on your browser:

```
Cookie received. Please close this page.
```

## 7. Configuring Additional Policies

Once you have logged in, you can now setup additional policies for other users.

E.g 1. The following command creates a policy with rules that allows an user to perform any action on any resource in dispatch. 

> **NOTE:** If using github as Identity Provider, please use *user's email* (not user name in github) as subject during policy creation.

```bash
dispatch iam create policy east-devops-policy-1 --subject <abc@example.com> --action "*" --resource "*"
```

To check the created policy content:
```bash
$ dispatch iam get policy east-devops-policy-1 --wide
          NAME         |         CREATED DATE         |             RULES
----------------------------------------------------------------------------------
  east-devops-policy-1 | Sat Jan  1 10:17:16 PST 0000 | {
                       |                              |   "actions": [
                       |                              |     "*"
                       |                              |   ],
                       |                              |   "resources": [
                       |                              |     "*"
                       |                              |   ],
                       |                              |   "subjects": [
                       |                              |     "abc@example.com"
                       |                              |   ]
                       |                              | }
```

The following `action` verbs are supported in a policy:

- `get`
- `create`
- `update`
- `delete`

The following `resource` types are supported in a policy:

- `api`
- `baseimage`
- `event`
- `function`
- `iam`
- `image`
- `runs`
- `secret`
- `service`
- `subscription`

E.g. 2. You can restrict a user to read-only operations on certain resources in dispatch with
```bash
dispatch iam create policy east-ro-policy-1 --subject <xyz@example.com> --action "get" --resource "function,runs"
```

## 8. Logout of Dispatch
To logout, enter the following:
```bash
dispatch logout
```


