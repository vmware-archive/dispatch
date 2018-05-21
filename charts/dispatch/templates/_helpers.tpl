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
    {{ .Values.global.ingress.annotationsPrefix }}/auth-url: "http://{{ .Release.Name }}-identity-manager.{{ .Release.Namespace }}.svc.cluster.local/v1/iam/auth"
    {{ .Values.global.ingress.annotationsPrefix }}/auth-response-headers: "{{ .Values.global.ingress.responseHeaders }}"
    {{ .Values.global.ingress.annotationsPrefix }}/configuration-snippet: |
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
        - {{ default .Values.global.host .Values.ingress.host }}
      {{- end -}}
  {{- end -}}
{{- end -}}
{{- end -}}
