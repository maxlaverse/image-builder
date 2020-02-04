# DockerIgnore *
# DockerIgnore !go.*
FROM {{ExternalImage (Concat "golang:" (Parameter "goVersion" "1.13"))}}
WORKDIR /build
COPY ./go.mod ./go.sum ./
RUN go mod download