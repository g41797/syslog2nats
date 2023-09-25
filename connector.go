package syslog2nats

import (
	"github.com/g41797/sputnik"
	"github.com/nats-io/nats.go"
)

const connectorConfName = "connector"

type brokerConnConfig struct {
	CONN_URL string
}

var _ sputnik.ServerConnector = &natsConnector{}

func NewConnector() sputnik.ServerConnector {
	return new(natsConnector)
}

type NatsConnection struct {
	Shared   bool
	NatsConn *nats.Conn
}

type natsConnector struct {
	sc *NatsConnection
}

func (c *natsConnector) Connect(cf sputnik.ConfFactory) (conn sputnik.ServerConnection, err error) {
	if c.IsConnected() {
		return c.sc, nil
	}

	var conf brokerConnConfig

	if err = cf(connectorConfName, &conf); err != nil {
		return nil, err
	}

	nc, err := nats.Connect(conf.CONN_URL)
	if err != nil {
		return nil, err
	}

	c.sc = &NatsConnection{true, nc}

	return c.sc, nil
}

func (c *natsConnector) IsConnected() bool {
	if c.sc == nil {
		return false
	}

	if c.sc.NatsConn.IsClosed() {
		return false
	}

	return true
}

func (c *natsConnector) Disconnect() {
	if c.sc.NatsConn != nil {
		// Because connector block is closed last
		// by sputnik, it's save to use Close() instead of Drain()
		c.sc.NatsConn.Close()
	}
	c.sc = nil
}
