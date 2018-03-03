---
layout: default
---
# OAuth Client Integration

Dispatch is a serverless framework that requires developers and end-users to authenticate themselves before they are
using the platform.

Dispatch integrates with 3rd party identification providers (IDP) in order to provide authentication.  The goal for
Dispatch is to provide integrations with a number of IDPs, including enterprise IDPs such is vIDM.

This document provides instructions for how to integrate specific IDPs into Dispatch.  This is generally a prerequisite
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


```bash
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
* Click ``Create``

You should see ``Client ID`` and ``Client Secret`` in the next page, they are the credentials you will use in the next
step.

For more detailed information visit Google's [Setting up an OAuth2 App page](https://developers.google.com/identity/protocols/OpenIDConnect#appsetup)

Edit dispatch's install config.yaml to add the information of the Identity Provider.

```bash
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

```bash
...
dispatch:
  oauth2Proxy:
    provider: oidc
    oidcIssuerURL: <OIDC Issuer URL>
    clientID: <client-id>
    clientSecret: <client-secret>
```


## 2. Create Cookie Secret (Optional)

Dispatch uses HTTP session cookies to keep track users. It is optional to encrypt the cookie sent to the end users, but it is highly recommended for security reasons.

To generate a random secret key:
```bash
$ python -c 'import os,base64; print base64.b64encode(os.urandom(16))'
YVBLBQXd4CZo1vnUTSM/3w==
```

Specify the cookie secret in the install config.yaml's `oauth2proxy` section
```bash
...
dispatch:
  oauth2Proxy:
    ....
    cookieSecret: YVBLBQXd4CZo1vnUTSM/3w==
```

## 3. Re-run Dispatch Install

Install/Update your Dispatch installation as normal, with
```bash
dispatch install -f config.yaml
```









