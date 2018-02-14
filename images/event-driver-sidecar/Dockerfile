FROM vmware/photon:2.0

ADD bin/event-driver-sidecar-linux /event-driver-sidecar
RUN chmod +x /event-driver-sidecar

ENTRYPOINT ["/event-driver-sidecar"]

