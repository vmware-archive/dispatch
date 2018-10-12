from photon:latest

RUN tdnf install -y \
    go \
    docker \
    git \
    tar \
    gzip

ENV GOPATH="/root/go"
RUN go get github.com/google/go-containerregistry/cmd/ko

RUN mkdir -p ${GOPATH}/src/github.com/knative && \
    cd ${GOPATH}/src/github.com/knative && \
    git clone https://github.com/knative/serving.git

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv kubectl /usr/local/bin/kubectl

RUN curl -OL https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-218.0.0-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-218.0.0-linux-x86_64.tar.gz && \
    mv google-cloud-sdk /usr/local/lib/google-cloud-sdk

RUN curl -OL https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz && \
    tar zxf helm-v2.11.0-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/helm

COPY setup_knative.sh /usr/local/bin/setup_knative
RUN chmod +x /usr/local/bin/setup_knative

ENV PATH="${GOPATH}/bin:/usr/local/lib/google-cloud-sdk/bin:${PATH}"
ENTRYPOINT ["setup_knative"]
CMD ["$@"]

