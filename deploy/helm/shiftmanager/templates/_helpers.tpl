{{/*
Expand the name of the chart.
*/}}
{{- define "shiftmanager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Fully qualified app name.
*/}}
{{- define "shiftmanager.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "shiftmanager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "shiftmanager.labels" -}}
helm.sh/chart: {{ include "shiftmanager.chart" . }}
{{ include "shiftmanager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "shiftmanager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "shiftmanager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
ServiceAccount name.
*/}}
{{- define "shiftmanager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "shiftmanager.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Image reference (tag falls back to the chart appVersion).
*/}}
{{- define "shiftmanager.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}

{{/*
Name of the Secret the app reads sensitive env from (existing or generated).
*/}}
{{- define "shiftmanager.secretName" -}}
{{- if .Values.existingSecret -}}
{{- .Values.existingSecret -}}
{{- else -}}
{{- include "shiftmanager.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Hostname of the bundled PostgreSQL primary service.
*/}}
{{- define "shiftmanager.postgresqlHost" -}}
{{- if .Values.postgresql.fullnameOverride -}}
{{- .Values.postgresql.fullnameOverride -}}
{{- else -}}
{{- printf "%s-postgresql" .Release.Name -}}
{{- end -}}
{{- end -}}

{{/*
Derived DATABASE_URL. Uses the bundled PostgreSQL when enabled, otherwise the
external DSN. Not used when existingSecret is set.
*/}}
{{- define "shiftmanager.databaseUrl" -}}
{{- if .Values.postgresql.enabled -}}
{{- $a := .Values.postgresql.auth -}}
{{- printf "postgres://%s:%s@%s:5432/%s?sslmode=disable" $a.username $a.password (include "shiftmanager.postgresqlHost" .) $a.database -}}
{{- else -}}
{{- required "database.external.url is required when postgresql.enabled=false (or set existingSecret)" .Values.database.external.url -}}
{{- end -}}
{{- end -}}

{{/*
Hostname of the bundled Redis master service.
*/}}
{{- define "shiftmanager.redisHost" -}}
{{- if .Values.redis.fullnameOverride -}}
{{- printf "%s-master" .Values.redis.fullnameOverride -}}
{{- else -}}
{{- printf "%s-redis-master" .Release.Name -}}
{{- end -}}
{{- end -}}

{{/*
Derived REDIS_URL. Empty string means "use the in-memory cache".
Not used when existingSecret is set.
*/}}
{{- define "shiftmanager.redisUrl" -}}
{{- if .Values.redis.enabled -}}
{{- if .Values.redis.auth.enabled -}}
{{- printf "redis://:%s@%s:6379/0" .Values.redis.auth.password (include "shiftmanager.redisHost" .) -}}
{{- else -}}
{{- printf "redis://%s:6379/0" (include "shiftmanager.redisHost" .) -}}
{{- end -}}
{{- else -}}
{{- .Values.redis.external.url -}}
{{- end -}}
{{- end -}}
