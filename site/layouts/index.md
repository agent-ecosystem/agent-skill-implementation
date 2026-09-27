{{- /* No .Title heading: as in the HTML homepage, the hero headline in content
       supplies the single H1 and the site name lives in the chrome. */ -}}
{{ partial "llms-directive-md.html" . }}
{{/* The homepage hand-writes its hero and card grid as HTML in content, so the
     agent-facing markdown converts it back: otherwise the .md output is
     mostly CSS. */ -}}
{{ partial "agent-markdown.html" (dict "content" .RawContent "stripHTML" true) }}
