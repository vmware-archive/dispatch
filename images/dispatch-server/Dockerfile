FROM vmware/photon:2.0

ADD bin/dispatch-server-linux /dispatch-server
RUN chmod +x /dispatch-server

ENTRYPOINT ["/dispatch-server"]
CMD ["local", "--host=0.0.0.0"]
