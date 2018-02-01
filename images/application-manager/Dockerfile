FROM vmware/photon:2.0

ADD bin/application-manager-linux /application-manager
RUN chmod +x /application-manager

VOLUME ["/data/application-manager"]

ENTRYPOINT ["/application-manager"]
CMD ["--host=0.0.0.0", "--port=8080"]
