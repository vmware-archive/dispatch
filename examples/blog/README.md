## Prerequisite

### Dispatch
You will need a dispatch cluster deployed and configured, please follow the quick start instruction at [dispatch](https://github.com/vmware/dispatch) repository.

For minikube deployment, keep a note of the port of your dispatch host and dispatch api-gateway host. This is printed
out at the end of the dispatch installation, but you can easily fetch it via kubectl:

```
$ kubectl -n kong get service api-gateway-kongproxy
NAME                    TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)                      AGE
api-gateway-kongproxy   NodePort   10.101.3.80   <none>        80:32521/TCP,443:32696/TCP   49m

$ export DISPATCH_HOST=$(minikube ip)
$ export DISPATCH_API_URL=https://$DISPATCH_HOST:$(jq '."api-https-port"' $HOME/.dispatch/config.json)
```

> **Note:** We are setting the `DISPATCH_API_URL` to the host IP and the **https** port.

### Minio

Minio is a S3 compatible object store, we use it to store blog posts. You can deploy a minio server locally or use S3 instead.

Here we briefly provide a simple way to deploy minio in a kubernetes cluster with helm, if you have deployed dispatch, you should already have a kubernetes cluter and helm ready.
```
$ export MINIO_ACCESS_KEY=blogaccess
$ export MINIO_SECRET_KEY=blogsecret

$ helm install --name minio --namespace minio --set serviceType=NodePort,accessKey=$MINIO_ACCESS_KEY,secretKey=$MINIO_SECRET_KEY stable/minio
```
You need to keep a note of minio credentials, including ``hostname``, ``port``, ``accessKey``, ``secretKey``.

The minio ``hostname`` is simply `DISPATCH_HOST` if you installed into the same kubernetes cluster:
```
$ export MINIO_HOST=$DISPATCH_HOST
```

To fetch the ``port`` associated with the minio service:
```
$ kubectl get svc -n minio
NAME              TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
minio-minio-svc   NodePort   10.102.89.197   <none>        9000:31390/TCP   37s
$ export MINIO_PORT=$(kubectl get svc -n minio -o json | jq '.items[0].spec.ports[0].nodePort')
```

The ``accessKey`` and ``secretKey`` are stored in kubernetes secrets:
```
$ echo $(kubectl get secret minio-minio-user -n minio -o json | jq -r .data.accesskey | base64 -D)
blogaccess
```

## Register the image with Dispatch

If you haven't cloned the dispatch repository, now is the time:
```
$ git clone https://github.com/vmware/dispatch.git && cd dispatch
```

```
$ dispatch create base-image blog-webapp-base-image dispatchframework/nodejs-base:0.0.7 --language=nodejs
$ dispatch create image blog-webapp-image blog-webapp-base-image --runtime-deps examples/blog/package.json
```

Wait for both the base-image and image to be in the ``READY`` state:
```
$ dispatch get base-image
           NAME          |                            URL                             | STATUS |         CREATED DATE
--------------------------------------------------------------------------------------------------------------------------
  blog-webapp-base-image | 10.97.167.150:5000/dispatch-nodejs-blog-webapp:0.0.1-dev1  | READY  | Sat Jan  1 13:06:40 PST 0000
```

```
$ dispatch get images
        NAME        |                  URL                   |       BASEIMAGE        | STATUS |         CREATED DATE
--------------------------------------------------------------------------------------------------------------------------
  blog-webapp-image | 10.97.167.150:5000/d5590f2e9cd3:latest | blog-webapp-base-image | READY  | Sat Jan  1 13:39:12 PST 0000
```

## Secret

Create a ``secret.json`` file to store the minio credientials:
```
$ cat << EOF > secret.json
{
    "endPoint": "$MINIO_HOST",
    "port": "$MINIO_PORT",
    "accessKey": "$MINIO_ACCESS_KEY",
    "secretKey": "$MINIO_SECRET_KEY",
    "bucket": "blog"
}
EOF
$ cat secret.json
{
    "endPoint": "192.168.64.24",
    "port": "31390",
    "accessKey": "*****",
    "secretKey": "*****",
    "bucket": "blog"
}

```

Create dispatch secret with your just created ``secret.json``
```
$ dispatch create secret blog-webapp-secret secret.json
$ dispatch get secret blog-webapp-secret
Note: secret values are hidden, please use --all flag to get them

                   ID                  |        NAME        | CONTENT
--------------------------------------------------------------------
  dc899a09-0abc-11e8-976a-922ad6cb76e3 | blog-webapp-secret | <hidden>
```

## Upload the post.js as a Dispatch function

Create a dispatch function and associate it with your just created dispatch secret.

```
$ dispatch create function --image=blog-webapp-image post examples/blog --handler=./post.js --secret blog-webapp-secret
```

Wait for the function to be in the ``READY`` state:

```
$ dispatch get function post
  NAME |       IMAGE       | STATUS |         CREATED DATE
---------------------------------------------------------------
  post | blog-webapp-image | READY  | Sat Jan  1 13:43:57 PST 0000
```

## Milestone I: Execute the uploaded function with dispatch cli

Use dispatch cli to test if your images, secrets and functions are deployed correctly and ready to be used.

```
$ dispatch exec post --input '{"op":"add", "post":{"id":"126", "title":"helloworld", "content":"this is a content"}}' --wait
```
```
{
    ...
    "output": {
        "post": {
            "content": "this is a content",
            "id": "126",
            "title": "helloworld"
        }
    },
    ...
    "status": "READY"
}
```

Try these for yourself and see what they do.
```
$ dispatch exec post --input '{"op":"get", "post":"126"}' --wait
$ dispatch exec post --input '{"op":"update", "post":{"id":"126", "title":"nihao", "content":"nihao"}}' --wait
$ dispatch exec post --input '{"op":"list"}' --wait
$ dispatch exec post --input '{"op":"delete", "post":{"id":"126"}}' --wait
```

## Create APIs

APIs are used by the blog webapp client (an angular2.0 project)

Create an unauthenticated GET endpoint with path `/post/list` that executes the `post` function
```
$ dispatch create api list-post-api post --auth public -m GET --path /post/list --cors
```

Create an unauthenticated GET endpoint with path `/post/get` that executes the `post` function
```
$ dispatch create api get-post-api post --auth public -m GET  --path /post/get --cors
```

Create an unauthenticated POST endpoint with the path `/post/add` that executes the `post` function
```
$ dispatch create api add-post-api post --auth public -m POST  --path /post/add --cors
```

Create an unauthenticated PATCH endpoint with the path `/post/update` that executes the `post` function
```
$ dispatch create api update-post-api post --auth public -m PATCH --path /post/update --cors
```

Create an unauthenticated DELETE endpoint with the path `/post/delete` that executes the `post` function
```
$ dispatch create api delete-post-api post --auth public -m DELETE --path /post/delete --cors
```

The `post` function is responsible for determining the operation, parsing url parameters, and performing
the correct operation.

Check the status of the APIs:

```
$ dispatch get api
       NAME       | FUNCTION | PROTOCOL  | METHOD | DOMAIN |     PATH     |  AUTH  | STATUS | ENABLED
-------------------------------------------------------------------------------------------------------
  update-post-api | post     | http      | PATCH  |        | /post/update | public | READY  | true
                  |          | https     |        |        |              |        |        |
-------------------------------------------------------------------------------------------------------
  delete-post-api | post     | http      | DELETE |        | /post/delete | public | READY  | true
                  |          | https     |        |        |              |        |        |
-------------------------------------------------------------------------------------------------------
  list-post-api   | post     | http      | GET    |        | /post/list   | public | READY  | true
                  |          | https     |        |        |              |        |        |
-------------------------------------------------------------------------------------------------------
  get-post-api    | post     | http      | GET    |        | /post/get    | public | READY  | true
                  |          | https     |        |        |              |        |        |
-------------------------------------------------------------------------------------------------------
  add-post-api    | post     | http      | POST   |        | /post/add    | public | READY  | true
                  |          | https     |        |        |              |        |        |
-------------------------------------------------------------------------------------------------------
```

## Milestone II: Execute the function via API Gateway

In this milestone, you want to test if your function works well via dispatch api-gateway.

If your dispatch is locally deployed, in this step, you need the https port on which your dispatch api-gateway is hosted on. It should already be set as ``DISPATCH_API_URL``.

```
$ echo $DISPATCH_API_URL
```

Add a new blog post
```
$ curl -s -k -X POST ${DISPATCH_API_URL}/post/add -H "Content-Type: application/json" -d '{
    "op": "add",
    "post":{
        "id": "1234",
        "title": "foo",
        "content":"bar bar bar"
    }
}' | jq
```
```
{
  "post": {
    "content": "bar bar bar",
    "title": "foo",
    "id": "1234"
  }
}
```

Get the newly created blog post
```
$ curl -s -k -X GET ${DISPATCH_API_URL}/post/get?op=get\&post=1234 | jq
```
```
{
  "post": {
    "content": "bar bar bar",
    "title": "foo",
    "id": "1234"
  }
}
```

Get a list of blog posts
```
$ curl -s -k -X GET ${DISPATCH_API_URL}/post/list?op=list | jq
```
```
{
  "post": [
    {
      "content": "bar bar bar",
      "title": "foo",
      "id": "1234"
    },
    {
      "content": "this is a content",
      "title": "helloworld",
      "id": "126"
    }
  ]
}
```

Update a blog post
```
$ curl -s -k -X PATCH ${DISPATCH_API_URL}/post/update -H "Content-Type: application/json" -d '{
    "op": "update",
    "post":{
        "id": "1234",
        "title": "foo",
        "content":"foo foo foo"
    }
}' | jq
```
```
{
  "post": {
    "content": "foo foo foo",
    "title": "foo",
    "id": "1234"
  }
}
```

Delete a blog post
```
$ curl -s -k -X DELETE ${DISPATCH_API_URL}/post/delete -H "Content-Type: application/json" -d '{
    "op": "delete",
    "post": { "id": "1234"}
}' | jq
```
```
{
  "post": {
    "id": "1234"
  }
}
```

After completing this milestone, it is a good time to deploy a front-end web-app(UI), which provides a friendly user interface for your blog.

Please go ahead and check [dispatch-example-blog-web-client](https://github.com/seanhuxy/dispatch-example-blog-web-client)
