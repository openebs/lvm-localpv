{{/*
    This returns a "1" if the CRD is absent in the cluster
    Usage:
      {{- if (include "crdIsAbsent" (list <crd-name>)) -}}
      # CRD Yaml
      {{- end -}}
*/}}
{{- define "crdIsAbsent" -}}
    {{- $crdName := index . 0 -}}
    {{- $crd := lookup "apiextensions.k8s.io/v1" "CustomResourceDefinition" "" $crdName -}}
    {{- $output := "1" -}}
    {{- if $crd -}}
        {{- $output = "" -}}
    {{- end -}}

    {{- $output -}}
{{- end -}}

{{/*
    Adds extra annotations to CRDs. This targets two scenarios: preventing CRD recycling in case
    the chart is removed; and adding custom annotations.
    NOTE: This function assumes the element `metadata.annotations` already exists.
    Usage:
      {{- include "crds.extraAnnotations" .Values.csi.volumeSnapshots | nindent 4 }}
*/}}

{{- define "crds.extraAnnotations" -}}
{{- if .keep -}}
helm.sh/resource-policy: keep
{{ end }}
{{- with .annotations }}
  {{- toYaml . }}
{{- end }}
{{- end -}}