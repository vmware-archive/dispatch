var Minio = require('minio')

var bucketExistsOrCreate = function (client, bucket) {
    return new Promise((fulfill, reject) => {
        client.bucketExists(bucket, err => {
            if (err) {
                if (err.code != 'NoSuchBucket') {
                    return reject(`error checking bucket existence: ${err}`)
                }
                client.makeBucket(bucket, '', err => {
                    if (err) {
                        return reject(`error making bucket: ${err}`)
                    }
                    fulfill()
                })
            }
            fulfill()
        })
    })
}

module.exports = function (context, params) {

    let op = params["op"]
    let post = params["post"]

    // BEGIN of workaround:

    // importing secret from api-gateway is not supported
    // issue tracked at: https://github.com/vmware/dispatch/issues/171
    // should replace with the commented code when the issue is addressed
    let client = new Minio.Client({
        "endPoint": "192.168.99.102",
        "port": 31515,
        "secure": false,
        "accessKey": "AKIAIOSFODNN7EXAMPLE",
        "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    })
    let bucket = "post-bucket"
    // let client = new Minio.Client({
    //     "endPoint": context.secrets["endPoint"],
    //     "port": parseInt(context.secrets["port"]),
    //     "secure": false,
    //     "accessKey": context.secrets["accessKey"],
    //     "secretKey": context.secrets["secretKey"]
    // })
    // let bucket = context.secrets["bucket"]

    // END of workaround

    console.log(`minio input params ${JSON.stringify(params)}`)

    operators = {
        "add": addPost,
        "update": updatePost,
        "list": listPosts,
        "get": getPost,
        "delete": deletePost
    }
    if (operators[op] == undefined) {
        return { error: "invalid operator" }
    }

    return new Promise((fulfill, reject) => {
        bucketExistsOrCreate(client, bucket).then(() => {
            console.log(`bucket existence check passed`)

            operators[op](client, bucket, post).then((post) => {
                console.log(`${op} post ${JSON.stringify(post)}`)
                fulfill({ post: post })
            }).catch((err) => {
                console.log(err)
                reject({ error: err })
            })
        }).catch(err => {
            console.log(err)
            reject({ error: err })
        })
    })
};

var _putPost = function (client, bucket, post) {

    return new Promise((fulfill, reject) => {
        client.putObject(bucket, post.id, JSON.stringify(post), (err, etag) => {
            console.log(`putObject: put post=${JSON.stringify(post)}`)
            if (err) {
                return reject(`error adding post: ${err}`)
            }
            return fulfill(post)
        })
    })
}

var addPost = function (client, bucket, post) {

    return new Promise((fulfill, reject) => {
        getPost(client, bucket, post.id).then((post, err) => {
            return reject(`error adding post: already exists`)
        }).catch((err) => {
            return fulfill(_putPost(client, bucket, post))
        })
    })
}

var updatePost = function (client, bucket, post) {

    return new Promise((fulfill, reject) => {
        getPost(client, bucket, post.id).then((_, err) => {
            return fulfill(_putPost(client, bucket, post))
        }).catch((err) => {
            return reject(`error updating: no such post`)
        })
    })
}

var getPost = function (client, bucket, postId) {
    return new Promise((fulfill, reject) => {
        client.getObject(bucket, postId, (err, stream) => {
            data = ""
            if (err) {
                return reject(`error getting object: ${err}`)
            }
            stream.on('error', err => {
                err = `error streaming object: ${err}`
                reject(err)
            })
            stream.on('data', chunk => {
                data += chunk
            })
            stream.on('end', () => {
                fulfill(JSON.parse(data))
            })
        })
    })
}

var listPosts = function (client, bucket) {

    return new Promise((fulfill, reject) => {
        var promises = []
        var stream = client.listObjects(bucket, '', false)
            .on('data', obj => {
                promises.push(getPost(client, bucket, obj.name))
            }).on('error', err => {
                reject(`error listing posts: ${err}`)
            }).on('end', () => {
                fulfill(Promise.all(promises))
            })
    })
}

var deletePost = function (client, bucket, post) {
    return new Promise((fulfill, reject) => {
        client.removeObject(bucket, post.id, (err) => {
            if (err) {
                return reject(`error removing post: ${err}`)
            }
            fulfill(post)
        })
    })
}