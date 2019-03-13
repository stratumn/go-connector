FROM golang:1.12-alpine

RUN addgroup -S -g 9999 connector
RUN adduser -H -D -u 9999 -G connector connector

RUN mkdir -p /usr/local/var/connector
RUN chown connector:connector /usr/local/var/connector/
RUN chmod 0700 /usr/local/var/connector

RUN mkdir -p /usr/local/bin
COPY dist/linux-amd64/connector /usr/local/bin/
COPY config.core.toml /usr/local/var/connector

USER connector

WORKDIR /usr/local/var/connector

EXPOSE 8903 8904 8905 8906
VOLUME [ "/usr/local/var/connector" ]

ENTRYPOINT [ "/usr/local/bin/connector" ]