///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

const defaultInstallConfigYaml = `
HelmRepositoryURL: https://s3-us-west-2.amazonaws.com/dispatch-charts
ingress:
  chart:
    chart: nginx-ingress
    namespace: kube-system
    release: ingress
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
  hostname: api.dev.dispatch.vmware.com
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
  hostname: dev.dispatch.vmware.com
  port: 443
  tls:
    secretName: dispatch-tls
  image:
    host: vmware
    tag: v0.1.2
  database: postgres
  debug: true
  trace: true
  persistData: false
  openfaasRepository:
    host: <host>
    username: <username>
    email: <email>@vmware.com
    password: <password>
  oauth2Proxy:
    clientID: <client-id>
    clientSecret: <client-secret>
    cookieSecret:
`
