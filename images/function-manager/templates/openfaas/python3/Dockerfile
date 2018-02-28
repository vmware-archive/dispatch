FROM vmware/dispatch-openfaas-watchdog:revbf667b8 AS watchdog
FROM {{ .DockerURL }}
COPY --from=watchdog /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

WORKDIR /root/

COPY index.py .
RUN pip3 install -U setuptools

RUN mkdir function && touch function/__init__.py
COPY {{ .FunctionFile }} function/handler.py

ENV fprocess="python3 index.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]