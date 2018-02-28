FROM vmware/dispatch-openfaas-watchdog:revbf667b8 AS watchdog
FROM {{ .DockerURL }}
COPY --from=watchdog /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

RUN mkdir -p /root/openfaas/function
WORKDIR /root/openfaas

# Function
COPY index.js .
COPY package.json .
RUN npm install

COPY {{ .FunctionFile }} function/func.js

ENV cgi_headers="true"

ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
