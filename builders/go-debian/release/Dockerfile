FROM {{BuilderStage "modules"}} AS builder
COPY . .
{{if HasFile "go.mod"}}
RUN {{Parameter "goBuildCmd" (Concat "go build -mod=readonly -o " (Parameter "binary"))}}
{{else}}
RUN {{Parameter "goBuildCmd" (Concat "go build -o " (Parameter "binary"))}}
{{end}}

FROM {{BuilderStage "with-packages"}}
{{if Parameter "assetsCopy"}}
{{range $key, $val := (Parameter "assetsCopy")}}
COPY --from={{BuilderStage "assets"}} /assets/{{$key}} {{$val}}
{{end}}
{{end}}
COPY --from=builder /build/{{MandatoryParameter "binary"}} /bin/{{MandatoryParameter "binary"}}
