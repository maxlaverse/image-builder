FROM {{BuilderStage "modules"}} AS builder
COPY . .
RUN {{Parameter "goBuildCmd" "go build"}}
FROM {{BuilderStage "with-packages"}}
COPY --from=builder /build/{{MandatoryParameter "binary"}} /bin/{{MandatoryParameter "binary"}}