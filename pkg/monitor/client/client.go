package client

import (
	"net/url"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/protobuf"
)

// Client ..
type Client struct {
	pb.MonitorClient
}

// NewClient ..
func NewClient(endpoint *url.URL) *Client {
	log.WithField("endpoint", endpoint.String()).Debug("establishing gRPC connection to the server..")

	targetAddress := endpoint.String()
	if endpoint.Scheme != "unix" {
		// Drop the schema and just use "host:port" if we're dealing with local addresses
		targetAddress = endpoint.Host
	}

	conn, err := grpc.NewClient(
		targetAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.WithField("endpoint", endpoint.String()).WithField("error", err).Fatal("could not connect to the server")
	}

	log.Debug("gRPC connection established")

	return &Client{
		MonitorClient: pb.NewMonitorClient(conn),
	}
}
