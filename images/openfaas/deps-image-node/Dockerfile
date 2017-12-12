FROM vmware/dispatch-openfaas-watchdog:revbf667b8
FROM vmware/photon:2.0
COPY --from=0 /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

RUN tdnf install -y nodejs-7.7.4-4.ph2

WORKDIR /root/

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

# Wrapper/boot-strapper
COPY package.json .
RUN npm i

# Function
COPY index.js .
RUN mkdir -p ./root/function

COPY function function

ENV cgi_headers="true"

ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
