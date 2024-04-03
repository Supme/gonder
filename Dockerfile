FROM golang:1.22 as builder
WORKDIR /app
ADD . /app
#RUN go generate
RUN --mount=type=cache,target="/tmp/.cache/go-build" GOCACHE=/tmp/.cache/go-build go build -ldflags "-s -w -linkmode external -extldflags -static -X gonder/models.AppVersion=`git describe --tags --abbrev=0` -X gonder/models.AppCommit=`git describe --always` -X gonder/models.AppDate=`date -u +%FT%TZ`" -o start .
RUN cd /app/cert && \
    openssl req -x509 -sha256 -nodes -days 3650 -newkey rsa:4096 -keyout server.key -out server.pem -subj "/C=RU/ST=Moscow/L=Moscow/O=Supme/OU=Gonder/CN=gonder.supme.ru"

FROM alpine:3.18 as production
LABEL maintainer="Supme <supme@gmail.com>"
WORKDIR /app
COPY --from=builder /app/start /app/
COPY --from=builder /app/cert /app/cert
COPY --from=builder /app/dist_config.toml /app/
RUN mkdir /app/files
EXPOSE 8080 7777
ENTRYPOINT ["./start"]
