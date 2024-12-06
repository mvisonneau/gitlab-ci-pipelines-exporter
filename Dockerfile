##
# BUILD CONTAINER
##

FROM alpine:3.21@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45 as certs

RUN \
apk add --no-cache ca-certificates

##
# RELEASE CONTAINER
##

FROM busybox:1.37-glibc@sha256:0911a82648a678cf004e61d4ac4e809e087ae22e7fcc6338861d6f6ac8277099

WORKDIR /

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gitlab-ci-pipelines-exporter /usr/local/bin/

# Run as nobody user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD ["run"]
