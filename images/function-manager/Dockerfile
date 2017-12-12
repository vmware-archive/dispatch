FROM vmware/photon:2.0

ADD bin/function-manager-linux /function-manager
RUN chmod +x /function-manager

VOLUME ["/data/function-manager"]

ENTRYPOINT ["/function-manager"]
CMD ["--db-file=/data/function-manager/db.bolt", "--host=0.0.0.0", "--port=8000"]