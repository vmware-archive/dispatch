FROM photon:1.0

RUN tdnf install -y nodejs-7.7.4-1.ph1

RUN curl -f -sSL https://github.com/openfaas/faas/releases/download/0.6.7/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

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
