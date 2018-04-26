---
layout: default
---
# Authentication in Dispatch

Dispatch is a serverless framework that requires developers and end-users to authenticate themselves before they are
using the platform.

Users in Dispatch are managed by an external Identity Provider (IDP) like Github, Google Identity Platform or other enterprise [OpenID Connect](https://openid.net/connect/) (OIDC) providers like vIDM.
OpenID Connect enhances OAuth 2.0 authorization protocol workflow to support authentication.
When users login to dispatch they will be redirected to the configured OIDC provider for authentication.

This document provides instructions for how to integrate specific IDP's into Dispatch.  This is generally a prerequisite
to setting up a Dispatch deployment.

## 1. Create An OAuth Client App with your Identity Provider


> **NOTE:** You will need to setup a different `OAuth App` for every dispatch deployment with a different Hostname/IP.

### Using Github

Login to your Github Account and go to [Github developer portal](https://github.com/settings/developers)

Click ``New OAuth App`` Button to create a new client app.

- Application name: for your reference only, e.g. ``dev-dispatch-app``
- Homepage URL: the hostname or IP address of your Dispatch deployment, e.g. ``https://dev.dispatch.vmware.com``
- Authorization callback URL: <homepage-url>/v1/iam/oauth2/callback, e.g. ``https://dev.dispatch.vmware.com/v1/iam/oauth2/callback``

Click ``Register application`` and now the client app is created

You should see ``Client ID`` and ``Client Secret`` in the next page, they are the credentials you will use in the next
step.

Edit dispatch's install config.yaml to add the information of the Identity Provider.


```yaml
...
dispatch:
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

Edit dispatch's install config.yaml to add the information of the Identity Provider.

```yaml
...
dispatch:
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
...
dispatch:
  oauth2Proxy:
    provider: oidc
    oidcIssuerURL: <OIDC Issuer URL>
    clientID: <client-id>
    clientSecret: <client-secret>
```


## 2. Create Cookie Secret (Optional)

Dispatch uses HTTP session cookies to keep track of users. It is optional to encrypt the cookie sent to the end users, but it is highly recommended for security reasons.

To generate a random secret key:
```bash
$ python -c 'import os,base64; print base64.b64encode(os.urandom(16))'
YVBLBQXd4CZo1vnUTSM/3w==
```

Specify the cookie secret in the install config.yaml's `oauth2proxy` section
```yaml
...
dispatch:
  oauth2Proxy:
    ....
    cookieSecret: YVBLBQXd4CZo1vnUTSM/3w==
```

## 3. Enable Bootstrap Mode

If you are enabling Authentication in dispatch for the first time, you will have to install it in the bootstrap mode.
In the bootstrap mode, the specified bootstrap user can configure the inital authorization policies. Without any authorization policies, even if the authentication is successful, users will be denied access to protected resources in dispatch.

You should always **disable** the bootstrap mode as soon as you have setup the required policies for an admin user.


> **Note:** If you have a running dispatch deployment with `skipAuth: true` in the dispatch `config.yaml`, you need to set it to `false` as part of this step for the bootstrap mode to work.

```yaml
...
dispatch:
  # Ensure skipAuth is unset or false (default is false)
  skipAuth: false
  # This must be a valid user managed by your identity provider
  bootstrapUser: xyz@example.com

```

## 4. Update Dispatch

Install/Update your Dispatch installation as normal, with
```bash
dispatch install -f config.yaml
```
> **TIP:** Dispatch install command can be used to update your running dispatch deployment.

If you already have a Dispatch deployment, you can also use *manage* subcommand to enable the bootstrap mode:
```bash
dispatch manage --enable-bootstrap-mode --bootstrap-user <BOOTSTRAP_USER> -f config.yaml
```
> **NOTE:** Please wait about 30 seconds for the changes to be applied.

## 5. Login to Dispatch

Login to dispatch with
```bash
dispatch login
```

You will now be redirected to your configured Identity Provider for authentication on a browser.

Sign-in to your Identity Provider as the `bootrstrapUser` that you configured in the previous step. Upon successful authentication, you should see the following response on your browser:

```
Cookie received. Please close this page.
```

## 6. Configure Policies

Once you have logged in as the `bootstrapUser`, you should setup the initial authorization policies for an admin user and then disable the bootstrap mode.

Execute the following command to create a policy with rules that allows the admin user to perform any action on any resource in dispatch. Note: replace the `<BOOTSTRAP_USER>` with an user account that is managed by your identity provider.

```bash
dispatch iam create policy default-admin-policy --subject <BOOTSTRAP_USER> --action "*" --resource "*"
```
> **NOTE:** If using github as Identity Provider, please use *user's email* (not user name in github) as subject during policy creation.

To check the created policy content:
```bash
$ dispatch iam get policy default-admin-policy --wide
          NAME         |         CREATED DATE         |             RULES
----------------------------------------------------------------------------------
  default-admin-policy | Sat Jan  1 10:17:16 PST 0000 | {
                       |                              |   "actions": [
                       |                              |     "*"
                       |                              |   ],
                       |                              |   "resources": [
                       |                              |     "*"
                       |                              |   ],
                       |                              |   "subjects": [
                       |                              |     "xyz@example.com"
                       |                              |   ]
                       |                              | }
```

To verify that the admin policy is in effect, logout and login as the admin user and run any privileged dispatch CLI commands. To logout, enter the following:
```bash
dispatch logout
```

## 7. Disable Bootstrap Mode [**Important!!!**]

The bootstrap mode is only to setup the initial authorization policies and must be disabled as soon as you have created an admin policy. To disable the bootstrap mode, simply use *manage* subcommand:
```bash
dispatch manage --disable-bootstrap-mode -f config.yaml
```
> **NOTE:** Please wait about 30 seconds for the changes to be applied.







