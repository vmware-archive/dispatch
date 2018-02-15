{{- define "ingress_with_external_auth" -}}
{{- $ingress_enabled := default .Values.global.ingress.enabled .Values.ingress.enabled -}}
{{- if $ingress_enabled -}}
{{- $tls := default .Values.global.tls .Values.ingress.tls -}}
{{- $ingress_host := default .Values.global.host .Values.ingress.host -}}
{{- $ingress_annotations := default .Values.global.ingress.annotations .Values.ingress.annotations -}}
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
    kubernetes.io/ingress.class: "nginx"
    {{- range $key, $value := $ingress_annotations }}
    {{ $key }}: {{ $value | quote }}
    {{- end }}
    nginx.ingress.kubernetes.io/auth-signin: "https://$host/v1/iam/oauth2/start"
    nginx.ingress.kubernetes.io/auth-url: "http://{{ .Release.Name }}-identity-manager.{{ .Release.Namespace }}.svc.cluster.local/v1/iam/auth"
    nginx.ingress.kubernetes.io/configuration-snippet: |
      error_page 403 = @403.json;
spec:
  rules:
    - http:
        paths:
          - path: {{ .Values.ingress.path }}
            backend:
              serviceName: {{ include "fullname" . }}
              servicePort: {{ .Values.service.externalPort }}
      {{- if $ingress_host }}
      host: {{ $ingress_host }}
      {{- end -}}
  {{- if $tls.secretName }}
  tls:
    - secretName: {{ $tls.secretName }}
      {{- if $ingress_host }}
      hosts:
        - {{ default .Values.global.host .Values.ingress.host }}
      {{- end -}}
  {{- end -}}
{{- end -}}
{{- end -}}