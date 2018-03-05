FROM vmware/photon:2.0

ADD bin/event-sidecar-linux /event-sidecar
RUN chmod +x /event-sidecar

ENTRYPOINT ["/event-sidecar"]

