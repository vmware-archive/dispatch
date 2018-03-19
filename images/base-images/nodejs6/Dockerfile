FROM vmware/photon2:20180302

RUN tdnf install -y nodejs-8.3.0-1.ph2

ARG IMAGE_TEMPLATE=/image-template
ARG FUNCTION_TEMPLATE=/function-template

LABEL io.dispatchframework.imageTemplate="${IMAGE_TEMPLATE}" \
      io.dispatchframework.functionTemplate="${FUNCTION_TEMPLATE}"

COPY image-template ${IMAGE_TEMPLATE}/
COPY function-template ${FUNCTION_TEMPLATE}/

ENV FUNCTION_MODULE=/function.js PORT=8080
EXPOSE ${PORT}

COPY function.js ${FUNCTION_MODULE}

COPY function-server /function-server/
RUN cd /function-server; npm install

CMD ["node", "/function-server/server.js"]
