# DockerIgnore *
# DockerIgnore !Gemfile*
FROM {{BuilderStage "base"}}
ENV GEM_HOME=/gems
RUN apt-get install -y build-essential {{if Parameter "buildPackages"}}{{range $val := (Parameter "buildPackages")}}{{$val}} {{end}}{{end}} \
  {{if Parameter "runtimePackages"}}{{range $val := (Parameter "runtimePackages")}}{{$val}} {{end}}{{end}}
COPY ./Gemfile ./Gemfile.lock ./
RUN bundle install --without development test