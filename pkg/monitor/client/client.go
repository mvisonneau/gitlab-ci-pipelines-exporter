package client

import (
	"context"
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
func NewClient(ctx context.Context, endpoint *url.URL) *Client {
	log.WithField("endpoint", endpoint.String()).Debug("establishing gRPC connection to the server..")

	conn, err := grpc.DialContext(
		ctx,
		endpoint.String(),
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
