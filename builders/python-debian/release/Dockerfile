FROM {{BuilderStage "python-packages"}} AS python-packages

FROM {{BuilderStage "system-packages"}}
COPY --from=python-packages /usr/local/lib/python{{MandatoryParameter "pythonVersion"}}/site-packages/ /usr/local/lib/python{{MandatoryParameter "pythonVersion"}}/site-packages/
COPY . /app