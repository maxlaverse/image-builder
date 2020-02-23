# DockerIgnore *
# DockerIgnore !Gemfile*
FROM {{BuilderStage "cache-packages"}}
ENV GEM_HOME=/gems
RUN apt-get install -y build-essential {{if Parameter "buildPackages"}}{{range $val := (Parameter "buildPackages")}}{{$val}} {{end}}{{end}}
COPY ./Gemfile ./Gemfile.lock ./
RUN bundle install --without development test