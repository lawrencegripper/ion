{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{ printf "dispatcher-%s" .Values.module.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- printf "%s-%s" (.Chart.Name | lower) .Values.module.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
