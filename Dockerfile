FROM golang:1.6

ADD ./scripts/bootstrap /scripts/bootstrap
RUN /scripts/bootstrap
