##
# BUILD CONTAINER
##

FROM alpine:3.22@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715 as certs

RUN \
apk add --no-cache ca-certificates

##
# RELEASE CONTAINER
##

FROM busybox:1.37-glibc@sha256:210ce53959959e79523b8cb0f0bb1cf1c49bf9747cdedb47db1cf0db8e642f61

WORKDIR /

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gitlab-ci-pipelines-exporter /usr/local/bin/

# Run as nobody user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD ["run"]
