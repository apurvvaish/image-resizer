{{- define "image-resizer.name" -}}
image-resizer
{{- end -}}

{{- define "image-resizer.fullname" -}}
{{ include "image-resizer.name" . }}
{{- end -}}

{{- define "image-resizer.chart" -}}
{{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
