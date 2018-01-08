---
output: true
---
# OAuth Client Integration

Dispatch is a serverless framework that requires developers and end-users to authenticate themselves before they are
using the platform.

Dispatch integrates with 3rd party identification providiers (IDP) in order to provide authorization.  The goal for
Dispatch is to provide integrations with a number of IDPs, including enterprise IDPs such is vIDM.  Currently GitHub is
the only supported IDP.

This document provides instructions for how to integrate specific IDPs into Dispatch.  This is generally a prerequisite
to setting up a Dispatch deployment.

## Github

### 1. Create An OAuth Client App with your Github Account

Login to your Github Account and go to [Github developer portal](https://github.com/settings/developers)

Click ``New OAuth App`` Button to create a new client app.

- Application name: for your reference only, e.g. ``dev-dispatch-app``
- Homepage URL: the hostname for your Dispatch deployment, e.g. ``https://dev.dispatch.vmware.com``
- Authorization callback URL: <homepage-url>/oauth2, e.g. ``https://dev.dispatch.vmware.com/oauth2``

Click ``Register application`` and now the client app is created

You should see ``Client ID`` and ``Client Secret`` in the next page, they are the credientials you will use in the next
step.

> **NOTE:** You will need to setup a different `OAuth App` for every deployment with a different hostname.

### 2. Create Cookie Secret (Optional)

Dispatch uses HTTP session cookies to keep track users. It is optional to encrypt the cookie sent to the end users, but it is highly recommended for security reasons.

To generate a random secret key:
```
$ python -c 'import os,base64; print base64.b64encode(os.urandom(16))'
YVBLBQXd4CZo1vnUTSM/3w==
```

### 3. Import Client App Credientials into Dispatch Chart

Install/Update your Dispatch chart as normal, with
```
export DISPATCH_OAUTH_CLIENT_ID = <client-id>
export DISPATCH_OAUTH_CLIENT_SECRET = <client-secret>
export DISPATCH_OAUTH_COOKIE_SECRET = <cookie-secret>

# install
helm install charts/dispatch --name=dev-dispatch --namespace dispatch \
    <other configurations> ... \
    --set oauth2-proxy.app.clientID=$DISPATCH_OAUTH_CLIENT_ID  \
    --set oauth2-proxy.app.clientSecret=$DISPATCH_OAUTH_CLIENT_SECRET \
    --set oauth2-proxy.app.cookieSecret=$DISPATCH_OAUTH_COOKIE_SECRET \
    --set --debug

# upgrade
# install
helm upgrade dev-dispatch charts/dispatch --namespace dispatch \
    <other configurations> ... \
    --set oauth2-proxy.app.clientID=$DISPATCH_OAUTH_CLIENT_ID  \
    --set oauth2-proxy.app.clientSecret=$DISPATCH_OAUTH_CLIENT_SECRET \
    --set oauth2-proxy.app.cookieSecret=$DISPATCH_OAUTH_COOKIE_SECRET \
    --set --debug
```









