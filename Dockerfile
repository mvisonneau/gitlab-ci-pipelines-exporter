##
# BUILD CONTAINER
##

FROM goreleaser/goreleaser:v0.112.2 as builder

WORKDIR /build

COPY Makefile .
RUN \
apk add --no-cache make ;\
make setup

COPY . .
RUN \
make build

##
# RELEASE CONTAINER
##

FROM busybox:1.31-glibc

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dist/gitlab-ci-pipelines-exporter_linux_amd64/gitlab-ci-pipelines-exporter /usr/local/bin/

# Run as nobody user
USER 65534

ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD [""]
