## Prerequisite

### Dispatch
You will need a dispatch cluster deployed and configured, please follow the quick start instruction at [dispatch](https://github.com/vmware/dispatch) repository.

You will need the url of your dispatch installation.
```
DISPATCH_HOST=<your-dispatch-host>
DISPATCH_API_HOST=<your-dispatch-api-host>
```

For minikube deployment, keep a note of the port of your dispatch host and dispatch api-gateway host.

### Minio

Minio is a S3 compatible object store, we use it to store blog posts. You can deploy a minio server locally or use S3 instead.

Here we briefly provide a simple way to deploy minio in a kubernetes cluster with helm, if you have deployed dispatch, you should already have a kubernetes cluter and helm ready.
```
helm upgrade minio --install --namespace minio --set serviceType=NodePort stable/minio
```
You need to keep a note of minio credentials, including ``hostname``, ``port``, ``accessKey``, ``secretKey``.

For minio deployed in kubernetes:

``hostname`` will be your kubernetes cluster hostname,

get ``port`` with
```kubectl get svc -n minio```,

get ``accessKey`` and ``secretKey`` with
```
kubectl get secret minio-minio-user -n minio -o yaml
```

## Build the image
```
export docker_user=<your-docker-username>
docker build -t ${docker_user}/dispatch-nodejs6-blog-webapp:0.0.1-dev1 ./base-image
docker push ${docker_user}/dispatch-nodejs6-blog-webapp:0.0.1-dev1
```

## Register the image with Dispatch

```
dispatch delete base-image blog-webapp-base-image
dispatch delete image blog-webapp-image
dispatch create base-image blog-webapp-base-image ${docker_user}/dispatch-nodejs6-blog-webapp:0.0.1-dev1 --language=nodejs6
dispatch create image blog-webapp-image blog-webapp-base-image
```

## Secret

Copy the ``secret.tmpl.json`` (which should be at the root folder of this example) to ``secret.json``, then replace with your minio credentials.
```
{
    "endPoint": "<minio-hostname>",
    "port": "<minio-port>",
    "accessKey": "<minio-access-key>",
    "secretKey": "<minio-secret-key>",
    "bucket": "<minio-bucket-name>"
}
```

Create dispatch secret with your just created ``secret.json``
```
dispatch delete secret blog-webapp-secret
dispatch create secret blog-webapp-secret secret.json
```

## Upload the post.js as a Dispatch function

Create a dispatch function and associate it with your just created dispatch secret.

```
dispatch delete function post
dispatch create function blog-webapp-image post post.js --secret blog-webapp-secret
```

## Milestone I: Execute the uploaded function with dispatch cli

Use dispatch cli to test if your images, secrets and functions are deployed correctly and ready to be used.

```
dispatch exec post --input '{"op":"add", "post":{"id":"126", "title":"helloworld", "content":"this is a content"}}' --wait
dispatch exec post --input '{"op":"get", "post":{"id":"126"}}' --wait
dispatch exec post --input '{"op":"update", "post":{"id":"126", "title":"nihao", "content":"nihao"}}' --wait
dispatch exec post --input '{"op":"list"}' --wait
dispatch exec post --input '{"op":"delete", "post":{"id":"126"}}' --wait
```

## Create APIs

APIs are used by the blog webapp client (an angular2.0 project)

```
dispatch delete api list-post-api
dispatch delete api get-post-api
dispatch delete api update-post-api
dispatch delete api add-post-api
dispatch delete api delete-post-api

dispatch create api list-post-api post --auth public -m GET --path /post/list --cors
dispatch create api get-post-api post --auth public -m GET  --path /post/get --cors
dispatch create api add-post-api post --auth public -m POST  --path /post/add --cors
dispatch create api update-post-api post --auth public -m PATCH --path /post/update --cors
dispatch create api delete-post-api post --auth public -m DELETE --path /post/delete --cors
```

## Milestone II: Execute the function via API Gateway

In this milestone, you want to test if your function works well via dispatch api-gateway.

If your dispatch is locally deployed, in this step, you need the https port on which your dispatch api-gateway is hosted on.

```
export DISPATCH_API_PORT=31841
export DISPATCH_API_URL=${DISPATCH_API_HOST}:${DISPATCH_API_PORT}

curl -X POST https://${DISPATCH_API_URL}/post/add -k -d '{
    "op": "add",
    "post":{
        "id": "1234",
        "title": "foo",
        "content":"bar bar bar"
    }
}'

curl -X GET https://${DISPATCH_API_URL}/post/get?op=get\&post=1234 -k
curl -X GET https://${DISPATCH_API_URL}/post/list?op=list -k

curl -X PATCH https://${DISPATCH_API_URL}/post/update -k -d '{
    "op": "update",
    "post":{
        "id": "1234",
        "title": "foo",
        "content":"foo foo foo"
    }
}'

curl -X DELETE https://${DISPATCH_API_URL}/post/delete -k -d '{
    "op": "delete",
    "post": { "id": "1234"}
}'
```

After completing this milestone, it is a good time to deploy a front-end web-app(UI), which provides a friendly user interface for your blog.

Please go ahead and check [dispatch-example-blog-web-client](https://github.com/seanhuxy/dispatch-example-blog-web-client)

