package rpc

import (
	"net/rpc"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor"
	log "github.com/sirupsen/logrus"
)

// Client ..
type Client struct {
	*rpc.Client
}

// NewClient ..
func NewClient() (c *Client) {
	c = &Client{}
	var err error
	c.Client, err = rpc.Dial("unix", SockAddr)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return
}

// Status ..
func (c *Client) Status() (s monitor.Status) {
	err := c.Call("Server.Status", "", &s)
	if err != nil {
		log.WithError(err).Fatal()
	}
	return
}

// Config ..
func (c *Client) Config() (s string) {
	err := c.Call("Server.Config", "", &s)
	if err != nil {
		log.WithError(err).Fatal()
	}
	return
}
