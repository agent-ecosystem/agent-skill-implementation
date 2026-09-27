{{- with .Title }}# {{ . }}

{{ end -}}
{{ partial "llms-directive-md.html" . }}
{{ partial "agent-markdown.html" (dict "content" .RawContent "stripHTML" false) }}
