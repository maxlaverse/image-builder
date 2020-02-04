# DockerIgnore **/.git
# DockerIgnore .image-builder-info
FROM {{BuilderStage "modules"}} AS builder
COPY . .
RUN go build

FROM {{BuilderStage "with-packages"}}
COPY --from=builder /build/{{MandatoryParameter "binary"}} /bin/{{MandatoryParameter "binary"}}