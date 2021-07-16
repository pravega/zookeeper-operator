{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "zookeeper-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "zookeeper-operator.fullname" -}}
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

{{/*
Common labels
*/}}
{{- define "zookeeper-operator.commonLabels" -}}
app.kubernetes.io/name: {{ include "zookeeper-operator.name" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
{{- end -}}

{{/*
Default sidecar template
*/}}
{{- define "chart.additionalSidecars"}}
{{ toYaml .Values.additionalSidecars }}
{{- end}}

{{/*
Default volume template
*/}}
{{- define "chart.additionalVolumes"}}
{{ toYaml .Values.additionalVolumes }}
{{- end}}

{{- define "crd.additionalPrinterColumns" }}
additionalPrinterColumns:
    - jsonPath: .spec.replicas
      description: The number of ZooKeeper servers in the ensemble
      name: Replicas
      type: integer
    - jsonPath: .status.readyReplicas
      description: The number of ZooKeeper servers in the ensemble that are in a Ready state
      name: Ready Replicas
      type: integer
    - jsonPath: .status.currentVersion
      description: The current Zookeeper version
      name: Version
      type: string
    - jsonPath: .spec.image.tag
      description: The desired Zookeeper version
      name: Desired Version
      type: string
    - jsonPath: .status.internalClientEndpoint
      description: Client endpoint internal to cluster network
      name: Internal Endpoint
      type: string
    - jsonPath: .status.externalClientEndpoint
      description: Client endpoint external to cluster network via LoadBalancer
      name: External Endpoint
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
{{- end }}

{{- define "crd.additionalPrinterColumnsV1Beta" }}
additionalPrinterColumns:
    - JSONPath: .spec.replicas
      description: The number of ZooKeeper servers in the ensemble
      name: Replicas
      type: integer
    - JSONPath: .status.readyReplicas
      description: The number of ZooKeeper servers in the ensemble that are in a Ready state
      name: Ready Replicas
      type: integer
    - JSONPath: .status.currentVersion
      description: The current Zookeeper version
      name: Version
      type: string
    - JSONPath: .spec.image.tag
      description: The desired Zookeeper version
      name: Desired Version
      type: string
    - JSONPath: .status.internalClientEndpoint
      description: Client endpoint internal to cluster network
      name: Internal Endpoint
      type: string
    - JSONPath: .status.externalClientEndpoint
      description: Client endpoint external to cluster network via LoadBalancer
      name: External Endpoint
      type: string
    - JSONPath: .metadata.creationTimestamp
      name: Age
      type: date
{{- end }}
