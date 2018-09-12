{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ingress_with_external_auth" -}}
{{- $ingress_enabled := .Values.ingress.enabled -}}
{{- if $ingress_enabled -}}
{{- $tls := .Values.ingress.tls -}}
{{- $ingress_host := .Values.ingress.host -}}
{{- $ingress_annotations := .Values.ingress.annotations -}}
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: {{ template "fullname" . }}
  labels:
    app: {{ template "name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
  annotations:
    kubernetes.io/ingress.class: "{{ .Values.ingress.class }}"
    {{- range $key, $value := $ingress_annotations }}
    {{ $key }}: {{ $value | quote }}
    {{- end }}
    # identity manager is not supported yet
    # {{ .Values.ingress.annotationsPrefix }}/auth-url: "http://{{ .Release.Name }}-identity-manager.{{ .Release.Namespace }}.svc.cluster.local/v1/iam/auth"
    # {{ .Values.ingress.annotationsPrefix }}/auth-response-headers: "{{ .Values.ingress.responseHeaders }}"
    {{ .Values.ingress.annotationsPrefix }}/configuration-snippet: |
      error_page 403 = @403.json;
      error_page 401 = @401.json;
spec:
  rules:
    - http:
        paths:
        {{- range $path := .Values.ingress.paths }}
          - path: {{ $path }}
            backend:
              serviceName: {{ include "fullname" $ }}
              servicePort: {{ $.Values.service.externalPort }}
        {{- end -}}
      {{- if $ingress_host }}
      host: {{ $ingress_host }}
      {{- end -}}
  {{- if $tls.secretName }}
  tls:
    - secretName: {{ $tls.secretName }}
      {{- if $ingress_host }}
      hosts:
        - {{ .Values.ingress.host }}
      {{- end -}}
  {{- end -}}
{{- end -}}
{{- end -}}
