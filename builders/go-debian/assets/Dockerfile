{{if Parameter "assetsCopy"}}
{{range $key, $val := (Parameter "assetsCopy")}}
# ContextInclude {{$key}}
{{end}}
{{end}}
FROM scratch
{{if Parameter "assetsCopy"}}
{{range $key, $val := (Parameter "assetsCopy")}}
COPY ./{{$key}} /assets/{{$val}}
{{end}}
{{end}}

