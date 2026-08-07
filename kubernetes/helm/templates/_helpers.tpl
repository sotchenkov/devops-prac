{{- define "homelab-backend.selectorLabels" -}}
app.kubernetes.io/name: app-backend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "homelab-backend.labels" -}}
{{ include "homelab-backend.selectorLabels" . }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}