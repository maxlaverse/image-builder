# DockerIgnore **/.git
# DockerIgnore .image-builder-info
FROM {{ExternalImage (Concat "debian:" (Parameter "osRelease" "buster"))}}
ENTRYPOINT ["/bin/{{MandatoryParameter "binary"}}"]
WORKDIR /app
{{if Parameter "runtimePackages"}}
RUN apt-get update && apt-get install -y {{range $val := (Parameter "runtimePackages")}}{{$val}} {{end}}
{{end}}