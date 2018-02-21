FROM vmware/photon:2.0

RUN tdnf install -y python3-3.6.1-9.ph2 \
    python3-pip-3.6.1-9.ph2

COPY requirements.txt   .
RUN pip3 install -U setuptools && pip3 install -r requirements.txt
