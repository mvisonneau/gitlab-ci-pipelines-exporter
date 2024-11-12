##
# BUILD CONTAINER
##

FROM alpine:3.20@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a as certs

RUN \
apk add --no-cache ca-certificates

##
# RELEASE CONTAINER
##

FROM busybox:1.37-glibc@sha256:3757a0aac2f46c8f8f96dae75b7f2b633d745252ef9d46bdce9c588a564cfc84

WORKDIR /

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gitlab-ci-pipelines-exporter /usr/local/bin/

# Run as nobody user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD ["run"]
