# UseBuilderContext
FROM {{BuilderStage "base"}}
RUN apt-get install -y libcurl4 {{if Parameter "runtimePackages"}}{{range $val := (Parameter "runtimePackages")}}{{$val}} {{end}}{{end}}