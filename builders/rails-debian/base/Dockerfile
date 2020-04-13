# UseBuilderContext
# ContextInclude rootfs/**/*
# FriendlyTag {{Parameter "osRelease"}}-ruby-{{File ".ruby-version"}}
# TagAlias {{Parameter "osRelease"}}-ruby-{{File ".ruby-version"}}
FROM {{ExternalImage (Concat "docker.io/library/debian:" (MandatoryParameter "osRelease"))}}
WORKDIR /app
ENV PATH=/root/.rbenv/bin:/root/.rbenv/shims:$PATH
RUN apt-get update -y \
  && apt-get install -y git curl procps libssl-dev libreadline-dev zlib1g-dev autoconf bison build-essential libyaml-dev libreadline-dev libncurses5-dev libffi-dev libgdbm-dev \
  && curl -sL https://github.com/rbenv/rbenv-installer/raw/master/bin/rbenv-installer | bash - \
  && rbenv install {{File ".ruby-version"}} \
  && rbenv global {{File ".ruby-version"}} \
  && gem install bundler -v 1.17.0 \
  && apt-get install -y dirmngr gnupg libcurl4-openssl-dev \
  && gem install passenger \
  && passenger-install-nginx-module --auto --languages ruby \
  && passenger-config install-standalone-runtime \
  && passenger start --runtime-check-only

COPY ./rootfs/entrypoint.sh /entrypoint.sh