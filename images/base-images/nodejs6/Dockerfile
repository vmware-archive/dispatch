FROM vmware/photon:2.0

RUN tdnf install -y nodejs-7.7.4-4.ph2

WORKDIR /root/

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

# Wrapper/boot-strapper
COPY package.json .
RUN npm i
