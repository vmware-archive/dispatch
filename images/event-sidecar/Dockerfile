FROM vmware/photon2:20180620

ADD bin/event-sidecar-linux /event-sidecar
RUN chmod +x /event-sidecar

ENTRYPOINT ["/event-sidecar"]

