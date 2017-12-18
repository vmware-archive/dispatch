


The Dispatch is a serverless platform that requires developers and end-users to authenticate themselves before they are using the platform.

We enable the Dispatch system admin to configure the identity provider they want to use.

Currently, GitHub is the only supported Identity Provider, the support for other providers is on the roadmap.

The wiki help the system admin to create an OAuth client application with their Github account, and import the client credientials to their Dispatch deployment.

### 1. Create An OAuth Client App with your Github Account

Login to your Github Account and go to [Github developer portal](https://github.com/settings/developers)

Click ``New OAuth App`` Button to create a new client app.

- Application name: for your reference only, e.g. ``dev-dispatch-app``
- Homepage URL: the hostname for your Dispatch deployment, e.g. ``https://dev.dispatch.vmware.com``
- Authorization callback URL: <homepage-url>/oauth2, e.g. ``https://dev.dispatch.vmware.com/oauth2``

Click ``Register application`` and now the client app is created

You should see ``Client ID`` and ``Client Secret`` in the next page, they are the credientials you will use in the next step.

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
helm install ./charts/dispatch --name=dev-dispatch --namespace dispatch \
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









