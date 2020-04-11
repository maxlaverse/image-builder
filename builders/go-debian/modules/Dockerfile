# DockerIgnore *
# DockerIgnore !go.*
FROM {{ExternalImage (Concat "docker.io/library/golang:" (Parameter "goVersion" "1.13"))}}
WORKDIR /build
{{if Parameter "goGetModules"}}
{{range $val := (Parameter "goGetModules")}}RUN go get {{$val}} {{end}}
{{end}}
COPY ./go.mod ./go.sum ./
RUN go mod download