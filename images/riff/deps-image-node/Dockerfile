FROM projectriff/node-function-invoker:0.0.3

FROM vmware/photon:2.0

RUN tdnf install -y nodejs-7.7.4-4.ph2

WORKDIR /usr/app
COPY --from=0 /usr/app/ .

# Turn down the verbosity level
ENV NPM_CONFIG_LOGLEVEL warn

RUN npm install --production

# Function
COPY function function
ENV FUNCTION_URI ./function/index.js

EXPOSE 8080
CMD ["node", "server.js"]
