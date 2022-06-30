##
# BUILD CONTAINER
##

FROM alpine:3.16.0 as certs

RUN \
apk add --no-cache ca-certificates

##
# RELEASE CONTAINER
##

FROM busybox:1.35.0-glibc

WORKDIR /

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gitlab-ci-pipelines-exporter /usr/local/bin/

# Run as nobody user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD ["run"]
