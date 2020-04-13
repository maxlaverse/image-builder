FROM {{BuilderStage "cache-gems-full"}} AS assets
COPY . /app
RUN DOMAIN_NAME=domain.com SECRET_KEY_BASE=test RAILS_ENV=production bundle exec rake assets:precompile

FROM {{BuilderStage "cache-gems-production-only"}} AS gems

FROM {{BuilderStage "cache-packages"}}
ENTRYPOINT ["/entrypoint.sh"]
ENV GEM_HOME=/app/gems
COPY --from=gems /gems/ /app/gems
COPY --from=gems /app/.bundle /app/.bundle
COPY . /app/
COPY --from=assets /app/public/ /app/public
# ContentHashIgnoreLine
LABEL gitSha {{GitCommitShort}}