FROM vmware/photon:2.0

COPY bin/function-manager-linux /function-manager
RUN chmod +x /function-manager

RUN mkdir -p /images/function-manager/templates
COPY images/function-manager/templates /images/function-manager/templates

VOLUME ["/data/function-manager"]

ENTRYPOINT ["/function-manager"]
CMD ["--db-file=/data/function-manager/db.bolt", "--host=0.0.0.0", "--port=8000"]