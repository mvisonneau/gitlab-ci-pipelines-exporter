##
# BUILD CONTAINER
##

FROM goreleaser/goreleaser:v0.123.3 as builder

WORKDIR /build

COPY . .
RUN \
apk add --no-cache make ca-certificates ;\
make build-linux-amd64

##
# RELEASE CONTAINER
##

FROM busybox:1.31-glibc

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dist/gitlab-ci-pipelines-exporter_linux_amd64/gitlab-ci-pipelines-exporter /usr/local/bin/

# Run as nobody user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD [""]
