{{/*
Expand the name of the chart.
*/}}
{{- define "kurator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "kurator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "kurator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kurator.labels" -}}
helm.sh/chart: {{ include "kurator.chart" . }}
{{ include "kurator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/name: kurator
{{- end }}

{{- define "kurator.selectorLabels" -}}
app.kubernetes.io/name: kurator
control-plane: controller-manager
{{- end }}

{{- define "kurator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (printf "%s-controller-manager" (include "kurator.fullname" .)) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "kurator.image" -}}
{{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
{{- end }}

{{- define "kurator.webhookServiceName" -}}
{{- printf "%s-webhook-service" (include "kurator.fullname" .) }}
{{- end }}

{{- define "kurator.servingCertName" -}}
{{- printf "%s-serving-cert" (include "kurator.fullname" .) }}
{{- end }}

{{- define "kurator.selfSignedIssuerName" -}}
{{- printf "%s-selfsigned-issuer" (include "kurator.fullname" .) }}
{{- end }}

{{- define "kurator.validatingWebhookConfigurationName" -}}
{{- printf "%s-validating-webhook-configuration" (include "kurator.fullname" .) }}
{{- end }}

{{- define "kurator.webhookCertInjectCAFrom" -}}
{{- printf "%s/%s" .Release.Namespace (include "kurator.servingCertName" .) }}
{{- end }}

{{- define "kurator.webhookServiceDNSNames" -}}
{{- $svc := include "kurator.webhookServiceName" . -}}
{{- printf "%s.%s.svc\n%s.%s.svc.cluster.local" $svc .Release.Namespace $svc .Release.Namespace }}
{{- end }}
