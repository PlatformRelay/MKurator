{{/*
Expand the name of the chart.
*/}}
{{- define "mkurator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "mkurator.fullname" -}}
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

{{- define "mkurator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "mkurator.labels" -}}
helm.sh/chart: {{ include "mkurator.chart" . }}
{{ include "mkurator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "mkurator.selectorLabels" -}}
app.kubernetes.io/name: mkurator
control-plane: controller-manager
{{- end }}

{{- define "mkurator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (printf "%s-controller-manager" (include "mkurator.fullname" .)) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "mkurator.image" -}}
{{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
{{- end }}

{{- define "mkurator.webhookServiceName" -}}
{{- printf "%s-webhook-service" (include "mkurator.fullname" .) }}
{{- end }}

{{- define "mkurator.servingCertName" -}}
{{- printf "%s-serving-cert" (include "mkurator.fullname" .) }}
{{- end }}

{{- define "mkurator.selfSignedIssuerName" -}}
{{- printf "%s-selfsigned-issuer" (include "mkurator.fullname" .) }}
{{- end }}

{{- define "mkurator.validatingWebhookConfigurationName" -}}
{{- printf "%s-validating-webhook-configuration" (include "mkurator.fullname" .) }}
{{- end }}

{{- define "mkurator.webhookCertInjectCAFrom" -}}
{{- printf "%s/%s" .Release.Namespace (include "mkurator.servingCertName" .) }}
{{- end }}

{{- define "mkurator.webhookServiceDNSNames" -}}
{{- $svc := include "mkurator.webhookServiceName" . -}}
{{- printf "%s.%s.svc\n%s.%s.svc.cluster.local" $svc .Release.Namespace $svc .Release.Namespace }}
{{- end }}
