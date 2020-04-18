# UseBuilderContext
FROM {{ExternalImage (Concat "docker.io/library/python:" (Parameter "pythonVersion") "-" (Parameter "osRelease"))}}
ENTRYPOINT ["/usr/local/bin/python","{{MandatoryParameter "startFile"}}"]
WORKDIR /app
{{if Parameter "runtimePackages"}}
RUN apt-get update && apt-get install -y {{range $val := (Parameter "runtimePackages")}}{{$val}} {{end}}
{{end}}