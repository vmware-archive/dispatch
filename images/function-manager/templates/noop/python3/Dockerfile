FROM {{ .DockerURL }}

WORKDIR /root/

COPY index.py .
RUN pip3 install -U setuptools

RUN mkdir function
RUN touch function/__init__.py
COPY {{ .FunctionFile }} function/handler.py
