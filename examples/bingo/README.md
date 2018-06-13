# Buzzword Bingo Slack Bot

In this example we create a fun little Slack bot that watches conversations in your channels and, if (or rather, when) someone uses 3 or more words like "serverless", "cloud" or "platform" (see the [source](bingo.js) for the currently supported list), it says:
> BINGO! ;)

## Create a Slack App

To simplify working with tokens (and to familiarize ourselves with what's next in the world of Slack apps) we're using a Slack developer preview apps in this demo: https://api.slack.com/slack-apps-preview

Once you have pressed that large green button, chosen the name and the workspace for the app, you can manage it.


### OAuth Scopes

In _OAuth & Permissions_ section, find _Scopes_ and add these:

- `channels:history` - needed to receive messages posted to channels 
- `chat:write` - needed to post messages

Save changes. Under _Your Workspace Token_, press "Install app to workspace". 


### Secrets

Create file `secret.json`:
```json
{
  "verificationToken": "",
  "oauthToken": ""
}
```

Once the app is installed, copy the _OAuth Access Token_ from app's Features -> OAuth & Permissions -> OAuth Access Token to the `"oauthToken"` field value in `secret.json`.

Slack also provides us with a verification token. We use it to make sure the requests are coming from Slack and are intended for our app. Go back to app's Settings -> Basic Information -> App Credentials and copy _Verfification Token_ to the `"verificationToken"` field value in `secret.json`.

Now, `secret.json` should look something like:

```json
{
  "verificationToken": "t0kEnTOk4n",
  "oauthToken": "xoxa-t0kenT0keNTOkEn"
}
```


### Event Subscription

To be able to act on posted chat messages, our bot needs to receive events from Slack. In _Event Subscriptions_ section flip ON the switch "Enable Events". Before you can enter Request URL, you'll need to create the API endpoint in Dispatch (we'll discuss that in a sec).

Let's just add `message.channels` by clicking `Add Workspace Event` button under Selected Events section and leave this page open for now.


## Create Dispatch Objects

Here are all the source files we need: [bingo.js](bingo.js), [package.json](package.json). 
One more thing: you need a working Dispatch installation and `dispatch` CLI on your computer (all right, that's two things). 


### Image

Every Dispatch function needs an image. Base images are intended to be basic runtimes for functions - without any additional libraries. Images are to contain the libraries.

```bash
## Register the image in Dispatch
dispatch create base-image js-base dispatchframework/nodejs-base:0.0.8
dispatch create image js-deps js-base --runtime-deps package.json
```

### Secret

We need to inject those secret values we've obtained while creating the Slack app. `secret.json` needs to have been created and populated.
```bash
dispatch create secret bingo secret.json
```

### Function

Let's make sure we have everything we need (image and secret) in READY state now:

```bash
dispatch get image js-deps
#
#     NAME    |                       URL                        |    BASEIMAGE    | STATUS |         CREATED DATE
#---------------------------------------------------------------------------------------------------------------------
# js-deps     | imikushin/dispatch-nodejs-js-deps:0.0.1-dev1     | js-deps-base    | READY  | Sat Jan  1 11:44:26 PST 0000
#
dispatch get secret bingo
#
#                   ID                  | NAME  | CONTENT
#-------------------------------------------------------
#  162f0dcb-f708-11e7-b15f-02f8d253c5b0 | bingo | <hidden>
```

Let's create the function:

```bash
dispatch create function --image=js-deps bingo . --handler=./bingo.js --secret=bingo
```

### API endpoint

Make sure the function is in READY state first:

```bash
dispatch get function bingo
#
#  NAME  |   IMAGE    | STATUS |         CREATED DATE
#---------------------------------------------------------
#  bingo | js-deps    | READY  | Sat Jan  1 11:47:53 PST 0000
```

Now the last piece of the puzzle, the API endpoint which will actually receive events from Slack.

```bash
dispatch create api bingo bingo -m POST -p /bingo
```

When the API endpoint is created, you should be able to see this:

```bash
dispatch get api bingo
#
#  NAME  | FUNCTION | PROTOCOL  | METHOD | DOMAIN |  PATH  |  AUTH  | STATUS | ENABLED
#---------------------------------------------------------------------------------------
#  bingo | bingo    | http      | POST   |        | /bingo | public | READY  | true
#        |          | https     |        |        |        |        |        |
#---------------------------------------------------------------------------------------
```

## Event Subscription URL

I've got Dispatch installed at https://dispatch.kops.bjung.net, so the API endpoint URL is `https://api.dispatch.kops.bjung.net/bingo` (`api.` is prepended to the DNS name and `/bingo` is the API endpoint path). 

So, paste the API endpoint URL into the Request URL field (on _Event Subscriptions_ page we have open). 

If everything worked out, you should see that the URL is "Verified" (Slack does it before sending any real traffic to the URL). The bot is now accepting events and ready to notify your team of buzzword abuse!


