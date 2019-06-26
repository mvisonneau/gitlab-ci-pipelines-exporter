##
# BUILD CONTAINER
##

FROM golang:1.12 as builder

WORKDIR /build

COPY Makefile .
RUN \
  make setup

COPY . .
RUN \
  make build-docker

##
# RELEASE CONTAINER
##

FROM alpine:3.9

WORKDIR /usr/local/bin

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /build/gitlab-ci-pipelines-exporter /usr/local/bin

# Run as nobody user
USER 65534

EXPOSE 8080/tcp
ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter", "-listen-address", ":8080"]
CMD [""]
