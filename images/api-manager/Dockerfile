FROM vmware/photon:2.0

ADD bin/api-manager-linux /api-manager
RUN chmod +x /api-manager

VOLUME ["/data/api-manager"]

ENTRYPOINT ["/api-manager"]
CMD ["--host=0.0.0.0", "--port=8080"]
