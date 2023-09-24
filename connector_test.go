package syslog2nats

import (
	"testing"

	_ "github.com/nats-io/nats-server/v2/server"
	_ "github.com/nats-io/nats.go"
	_ "github.com/nats-io/nats.go/jetstream"
)

func TestRunNATS(t *testing.T) {
	srv := RunBasicJetStreamServer(NATSPORT)
	if srv == nil {
		t.Fatalf("cannot start broker")
	}
	defer shutdownJSServerAndRemoveStorage(t, srv)
}

func TestConnect(t *testing.T) {
	srv := RunBasicJetStreamServer(NATSPORT)
	if srv == nil {
		t.Fatalf("cannot start broker")
	}
	defer shutdownJSServerAndRemoveStorage(t, srv)

	cntr := newConnector()

	defer cntr.Disconnect()

	_, err := cntr.Connect(ConfFact())

	if err != nil {
		t.Fatal(err)
	}

	connected := cntr.IsConnected()

	if !connected {
		t.Fatalf("should be connected")
	}

}
