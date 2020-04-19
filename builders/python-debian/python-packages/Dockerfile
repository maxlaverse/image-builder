# ContextInclude requirements.txt
FROM {{BuilderStage "system-packages"}}
WORKDIR /build
{{if Parameter "buildPackages"}}
RUN apt-get update && apt-get install -y {{range $val := (Parameter "buildPackages")}}{{$val}} {{end}}
{{end}}
COPY ./ ./
RUN pip3 install -r requirements.txt