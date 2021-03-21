# ContextInclude go.mod
# ContextInclude go.sum
FROM {{ExternalImage (Concat "docker.io/library/golang:" (Parameter "goVersion" "1.13"))}}
WORKDIR /build
{{if Parameter "buildPackages"}}
RUN apt-get update && apt-get install -y {{range $val := (Parameter "buildPackages")}}{{$val}} {{end}}
{{end}}
# Builder is a noop if there is no go.mod file
{{if HasFile "go.mod"}}
{{if Parameter "goGetModules"}}
{{range $val := (Parameter "goGetModules")}}RUN go get {{$val}} {{end}}
{{end}}
COPY ./go.mod ./go.sum ./
RUN go mod download
{{end}}
