##
# BUILD CONTAINER
##

FROM golang:1.10 as builder

WORKDIR /go/src/github.com/mvisonneau/gitlab-ci-pipelines-exporter

COPY Makefile .
RUN \
make setup

COPY . .
RUN \
make deps ;\
make build-docker

##
# RELEASE CONTAINER
##

FROM busybox:1.29

WORKDIR /usr/local/bin

COPY --from=builder /go/src/github.com/mvisonneau/gitlab-ci-pipelines-exporter/gitlab-ci-pipelines-exporter /usr/local/bin

EXPOSE 80/tcp
ENTRYPOINT ["/usr/local/bin/gitlab-ci-pipelines-exporter"]
CMD [""]
