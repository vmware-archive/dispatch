FROM photon:latest

ADD bin/image-manager-linux /image-manager
RUN chmod +x /image-manager

VOLUME ["/data/image-manager"]

ENTRYPOINT ["/image-manager"]
CMD ["--db-file=/data/image-manager/db.bolt", "--host=0.0.0.0", "--port=8000"]