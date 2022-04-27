package rpc

import (
	"net/rpc"
	"net/url"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor"
	log "github.com/sirupsen/logrus"
)

// Client ..
type Client struct {
	*rpc.Client
	serverAddress *url.URL
}

// NewClient ..
func NewClient(serverAddress *url.URL) (c *Client) {
	c = &Client{
		serverAddress: serverAddress,
	}

	var err error

	c.Client, err = rpc.Dial(c.serverAddress.Scheme, c.serverAddress.Host)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	return
}

// Status ..
func (c *Client) Status() (s monitor.Status) {
	if err := c.Call("Server.Status", "", &s); err != nil {
		log.WithError(err).Fatal()
	}

	return
}

// Config ..
func (c *Client) Config() (s string) {
	if err := c.Call("Server.Config", "", &s); err != nil {
		log.WithError(err).Fatal()
	}

	return
}
