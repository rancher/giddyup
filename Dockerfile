FROM golang:1.5

ADD ./scripts/bootstrap /scripts/bootstrap
RUN /scripts/bootstrap
