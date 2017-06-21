FROM debian:jessie
COPY giddyup /opt/rancher/bin/
COPY ./entrypoint.sh /entrypoint.sh
RUN ln -s /opt/rancher/bin/giddyup /usr/bin/giddyup

VOLUME /opt/rancher/bin

ENTRYPOINT ["/entrypoint.sh"]
CMD ["giddyup"]
