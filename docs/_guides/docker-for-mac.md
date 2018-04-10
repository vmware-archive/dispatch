---
layout: default
---
# Installing Dispatch on Kubernetes with Docker for Mac

Docker for Mac Edge comes with [Kubernetes support](https://docs.docker.com/docker-for-mac/kubernetes/) that you can use to run Dispatch. 

It's convenient to build Dispatch with Docker for Mac, because you don't have to switch to another docker environment to build Dispatch images. There is a caveat though: you can't run dispatch with the included local image registry. Docker daemon in Docker for Mac can't access it because it works in the network space of the Mac host. But fear not: there is a way around that! 

We are assuming you've got an Edge release of Docker for Mac with Kubernetes enabled and running. You have also downloaded or built the Dispatch CLI.


## Configure your Image Registry

### Local Image Registry

Run a standalone image registry:
```
docker run -d -p 5000:5000 --restart=always --name=registry registry
```

In Docker for Mac preferences:
- add `host.docker.internal:5000` to the list of insecure registries
- add `http://localhost:5000` to the list of registry mirrors

Use `host.docker.internal:5000` as the imageRegistry in dispatch install configuration file, `config.yaml`:

```yaml
apiGateway:
  host: 127.0.0.1
dispatch:
  host: 127.0.0.1
  debug: true
  skipAuth: true
  imageRegistry:
    name: host.docker.internal:5000
    insecure: true
```

### External Image Registry

Make sure you can login to your image registry:
```bash
$ docker login ${host}:${port}
```

Edit dispatch install configuration file, `config.yaml`:
```yaml
apiGateway:
  host: 127.0.0.1
dispatch:
  host: 127.0.0.1
  debug: true
  skipAuth: true
  imageRegistry:
    name: <host>:<port>
    username: <username>
    email: <user-email@domain.com>
    password: <password>
```

## Install Dispatch

Then, you can install Dispatch with the following command:
```bash
dispatch install -f ./config.yaml
```



