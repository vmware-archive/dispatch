FROM vmware/vs-openfaas-watchdog:rev2d268d9
FROM vmware/photon:2.0
COPY --from=0 /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

RUN tdnf install -y python3-3.6.1-9.ph2 \
    python3-pip-3.6.1-9.ph2

WORKDIR /root/

COPY index.py           .
COPY requirements.txt   .
RUN pip3 install -U setuptools
RUN pip3 install -r requirements.txt

COPY function           function

RUN touch ./function/__init__.py

WORKDIR /root/function/
COPY function/requirements.txt	.
RUN pip3 install -r requirements.txt

WORKDIR /root/

ENV fprocess="python3 index.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]