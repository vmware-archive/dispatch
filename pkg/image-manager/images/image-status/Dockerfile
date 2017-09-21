FROM docker:latest

ADD image_status.sh /image_status.sh
RUN chmod +x /image_status.sh

ENTRYPOINT ["/image_status.sh"]