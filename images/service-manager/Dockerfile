FROM vmware/photon:2.0

ADD bin/service-manager-linux /service-manager
RUN chmod +x /service-manager

VOLUME ["/data/service-manager"]

ENTRYPOINT ["/service-manager"]
CMD ["--db-file=/data/service-manager/db.bolt", "--host=0.0.0.0", "--port=8000"]