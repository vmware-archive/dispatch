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
  insecure: false
  #imageRegistry:
  #  name:
  #  username:
  #  email:
  #  password:
  oauth2Proxy:
    clientID: <client-id>
    clientSecret: <client-secret>
    cookieSecret:
`
