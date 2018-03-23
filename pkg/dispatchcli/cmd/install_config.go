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
  exposeService: false
kafka:
  chart:
    chart: kafka
    namespace: riff-system
    release: transport
    repo: https://riff-charts.storage.googleapis.com
    version: 0.0.1
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
dispatch:
  chart:
    chart: dispatch
    namespace: dispatch
    release: dispatch
  organization: dispatch
  host:
  port: 443
  tls:
    secretName: dispatch-tls
  image:
    host:
    tag:
  database: postgres
  debug: true
  trace: true
  persistData: false
  skipAuth: false
  bootstrapUser:
  insecure: false
  faas: openfaas
  #imageRegistry:
  #  name:
  #  username:
  #  email:
  #  password:
  oauth2Proxy:
    provider: github
    oidcIssuerURL:
    clientID: <client-id>
    clientSecret: <client-secret>
    cookieSecret:
`
