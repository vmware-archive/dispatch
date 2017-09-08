# Demo v1

This demo script is intended to exercise the core/basic functionality of the VMware serverless platform.  It is far from
full featured, but exercises some basic functionality of all core components (except for the API gateway).

```
# Login to vmware serverless and store credentials (token) in ~/.vs/config.json
$ vs login
Username: aladdin
Password:
Login Succeeded
# Create/register an existing base image
$ vs create base-image --name demo-python3-base registry.hub.docker.com/aladdin/base-python3
{
    "type": "base-image",
    "id": "d52971a1-fb24-4347-9450-8923c0995d20",
    "name": "demo-python3-base",
    "imageUri": "registry.hub.docker.com/aladdin/base-python3",
    "language": "python3",
    "defaultForLanguage": true,
    "state": "ready"
}
# Create a runnable image and set the status to published, which will immediately kick-off image building
# In this case we are not adding any additional packages
$ vs create image --name demo-python3-runtime demo-python3-base --status published --language python3
{
    "type": "image",
    "id": "e99086b9-7fc8-434b-9993-d0f11c6740b1",
    "name": "demo-python3-runtime",
    "status": {
        "published": true
    },
    "state": "created"
}
# Check image state... waiting for "ready"
$ vs get image demo-python3-runtime
{
    "type": "image",
    "id": "e99086b9-7fc8-434b-9993-d0f11c6740b1",
    "name": "demo-python3-runtime",
    "status": {
        "published": true
    },
    "state": "ready"
}
# Build a schema for our function
$ cat open-sesame.in.schema
{
    "title": "open-sesame.in",
    "type": "object",
    "properties": {
        "password": {
            "type": "string",
            "minLength": 8
        }
    },
    "required": ["password"]
}
# Build the function code
$ cat open-sesame.py
def main(args):
    secrets = args.get("__secrets", {})
    secret_password = secrets.get("password")
    if not secret_password:
        return {"greeting": "no password set, good-bye"}
    password = args.get("password")
    if password != secret_password:
        return {"greeting": "passwords do not match, good-bye"}
    return {"greeting": "welcome"}
# Create the function with associated schema
$ vs create function --name open-sesame --schema-in open-sesame.in.schema --image demo-python3-runtime open-sesame.py
{
    "type": "function",
    "id": "e116d555-cbb2-40db-ae18-966a93eba7cd",
    "name": "open-sesame",
    "runtimeImage": "e99086b9-7fc8-434b-9993-d0f11c6740b1",
    "schema": {
        "in": "68355b08-4c20-4d7d-8546-db61641a9153"
    },
    "data": "<base64encoded open-sesame.py>",
    "status": {
        "published": true
    }
    "state": "created"
}
# Check function state... waiting for "ready"
$ vs get function open-sesame
{
    "type": "function",
    "id": "e116d555-cbb2-40db-ae18-966a93eba7cd",
    "name": "open-sesame",
    "runtimeImage": "e99086b9-7fc8-434b-9993-d0f11c6740b1",
    "schema": {
        "in": "68355b08-4c20-4d7d-8546-db61641a9153"
    },
    "data": "<base64encoded open-sesame.py>",
    "status": {
        "published": true
    }
    "state": "ready"
}
# Build a secrets file
$ cat secrets.json
{
    "password": "0p3nSes4m3!"
}
# Create the secret
$ vs create secret --name demo-password secrets.json
{
    "type": "secret",
    "id": "1696fb98-e66f-4877-abb0-96543f01f47a",
    "name": "demo-password",
    "state": "ready"
}
# Execute the function and pass in the secrets... wait/block until the result is available
$ vs exec --secret demo-password --wait --params '{"password": "0p3nSes4m3!"}' open-sesame
{
    "id": "cdc9a192-9ca8-4a37-b831-de20930849f8",
    "stdout": "",
    "stderr": "",
    "result": {
        "greeting": "welcome"
    }
}
```