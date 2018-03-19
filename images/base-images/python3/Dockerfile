FROM vmware/photon2:20180302

RUN tdnf install -y python3-3.6.1-9.ph2 python3-pip-3.6.1-9.ph2 && pip3 install --upgrade pip setuptools

ARG IMAGE_TEMPLATE=/image-template
ARG FUNCTION_TEMPLATE=/function-template

LABEL io.dispatchframework.imageTemplate="${IMAGE_TEMPLATE}" \
      io.dispatchframework.functionTemplate="${FUNCTION_TEMPLATE}"

COPY image-template ${IMAGE_TEMPLATE}/
COPY function-template ${FUNCTION_TEMPLATE}/

ENV FUNCTION_MODULE=/function-server/function/handler.py PORT=8080
EXPOSE ${PORT}

COPY function-server /function-server/

WORKDIR /function-server
RUN pip install -r requirements.txt

CMD ["python3", "main.py"]
