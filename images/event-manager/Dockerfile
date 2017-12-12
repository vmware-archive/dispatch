FROM vmware/photon:2.0

ADD bin/event-manager-linux /event-manager
RUN chmod +x /event-manager

VOLUME ["/data/event-manager"]

ENTRYPOINT ["/event-manager"]
CMD ["--db-file=/data/event-manager/db.bolt", "--host=0.0.0.0", "--port=8000"]