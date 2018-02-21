FROM vmware/dispatch-openfaas-watchdog:revbf667b8 AS watchdog
FROM {{ .DockerURL }}
COPY --from=watchdog /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

WORKDIR /root/

# Function
COPY index.js .
RUN mkdir function

COPY {{ .FunctionFile }} function/func.js

ENV cgi_headers="true"

ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
