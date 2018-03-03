FROM vmware/photon:2.0

ADD bin/event-driver-linux /event-driver
RUN chmod +x /event-driver

ENTRYPOINT ["/event-driver"]
CMD ["vcenter"]
