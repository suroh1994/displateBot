FROM alpine

ARG USER=disbot
RUN adduser -D $USER

USER $USER

WORKDIR /
COPY ./displateBot /displateBot

ENTRYPOINT ["/displateBot"]