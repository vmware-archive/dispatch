FROM vmware/photon:2.0

ADD bin/identity-manager-linux /identity-manager
RUN chmod +x /identity-manager

VOLUME ["/data/identity-manager"]

ENTRYPOINT ["/identity-manager"]
CMD ["--config=/data/config/serverless.json", "--static-users=/data/config/user.dev.json", \
"--db-file=/data/identity-manager/db.bolt", "--host=0.0.0.0", "--port=8080"]
