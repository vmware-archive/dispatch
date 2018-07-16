///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

const defaultInstallConfigYaml = `
HelmRepositoryURL: https://s3-us-west-2.amazonaws.com/dispatch-charts
dockerRegistry:
  chart:
    chart: docker-registry
    namespace: dispatch
    release: docker-registry
    repo: https://kubernetes-charts.storage.googleapis.com
ingress:
  chart:
    chart: nginx-ingress
    namespace: kube-system
    release: ingress
    repo: https://kubernetes-charts.storage.googleapis.com
    opts:
      rbac.create: true
  serviceType: NodePort
postgresql:
  chart:
    chart: postgresql
    namespace: dispatch
    release: postgres
    repo: https://kubernetes-charts.storage.googleapis.com
    version: 0.8.5
  database: dispatch
  username: vmware
  password: dispatch
  host: postgresql
  port: 5432
  persistence: false
apiGateway:
  chart:
    chart: kong
    namespace: kong
    release: api-gateway
  serviceType: NodePort
  database: postgres
  host:
  tls:
    secretName: api-dispatch-tls
openfaas:
  chart:
    chart: openfaas
    namespace: openfaas
    release: openfaas
    repo: https://openfaas.github.io/faas-netes/
    version: 1.0.20
  exposeService: false
kafka:
  chart:
    chart: kafka
    namespace: dispatch
    release: transport
    repo: http://storage.googleapis.com/kubernetes-charts-incubator
  brokers:
  - transport-kafka.dispatch:9092
  zookeeperNodes:
  - zookeeper.zookeeper.svc.cluster.local:2181
rabbitmq:
  chart:
    chart: rabbitmq
    namespace: dispatch
    release: rabbitmq
    repo: https://kubernetes-charts.storage.googleapis.com
    version: 0.6.26
  username: dispatch
  password: dispatch
  host:
  persist: false
  port: 5672
riff:
  chart:
    chart: riff
    namespace: riff
    release: riff
    repo: https://riff-charts.storage.googleapis.com
    version: 0.0.4
    opts:
      create.rbac: true
      httpGateway.service.type: ClusterIP
kubeless:
  chart:
    chart: kubeless
    namespace: kubeless
    release: kubeless
    repo: http://storage.googleapis.com/kubernetes-charts-incubator
    version: 1.0.0
    opts:
      rbac.create: true
jaeger:
  chart:
    chart: jaeger
    namespace: jaeger
    release: jaeger
    repo: http://storage.googleapis.com/kubernetes-charts-incubator
    version: 0.3.7
    opts:
      cassandra.persistence.enabled: false
      cassandra.config.cluster_size: 1
      cassandra.config.seed_size: 1
      cassandra.resources.requests.cpu: 1
      cassandra.resources.requests.memory: 2Gi
  agent:
  enabled: false
zookeeper:
  chart:
    chart: zookeeper
    namespace: zookeeper
    release: zookeeper
    repo: http://storage.googleapis.com/kubernetes-charts-incubator
    version: 1.1.0
  location: zookeeper.zookeeper.svc.cluster.local
certManager:
  chart:
    repo: https://kubernetes-charts.storage.googleapis.com
    chart: cert-manager
    namespace: kube-system
    release: cert-manager
    version: 0.2.10
  enabled: false
letsEncrypt:
  chart:
    chart: certificate
    namespace: dispatch
    release: dispatch-certificate
  email: user@example.com
  staging: false
  dns:
    provider: clouddns
    clouddns:
      project: <GCP project ID>
      secretName: clouddns
      secretKey: service-account.json
    route53:
      accessKeyID: <aws access key ID>
      secretName: route53
      secretKey: secret-access-key
dispatch:
  chart:
    chart: dispatch
    namespace: dispatch
    release: dispatch
  host:
  port: 443
  tls:
    ca:
    insecure: false
    secretName: dispatch-tls
  image:
    host:
    tag:
  database: postgres
  debug: true
  trace: false
  persistData: false
  skipAuth: false
  faas: openfaas
  eventTransport: kafka
  #imageRegistry:
  #  name:
  #  username:
  #  email:
  #  password:
  imagePullSecret:
  service:
    catalog: k8sservicecatalog
    k8sservicecatalog:
      namespace: dispatch
  oauth2Proxy:
    provider: github
    oidcIssuerURL:
    clientID: <client-id>
    clientSecret: <client-secret>
    cookieSecret:
`
