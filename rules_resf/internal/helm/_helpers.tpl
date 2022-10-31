{{/*
Expand the name of the chart.
*/}}
{{- define "{NAME}.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "{NAME}.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "{NAME}.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Stage/environment the chart is being deployed to.
*/}}
{{- define "resf.stage" -}}
{{- .Values.stage | default "prod" }}
{{- end }}

{{/*
Long name of stage (prod -> production for example)
*/}}
{{- define "resf.longStage" -}}
{{- if eq .Values.stage "prod" }}
{{- "production" }}
{{- else if eq .Values.stage "dev" }}
{{- "development" }}
{{- else }}
{{- .Values.stage }}
{{- end }}
{{- end }}
