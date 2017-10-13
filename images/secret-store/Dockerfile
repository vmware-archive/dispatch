FROM photon:latest

ADD bin/secret-store-linux /secret-store
RUN chmod +x /secret-store

VOLUME ["/data/secret-store"]

ENTRYPOINT ["/secret-store"]
CMD ["--host=0.0.0.0", "--port=8080"]
