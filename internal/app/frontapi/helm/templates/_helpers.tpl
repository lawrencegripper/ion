{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{ printf "frontapi" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- printf "%s" (.Chart.Name | lower) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
