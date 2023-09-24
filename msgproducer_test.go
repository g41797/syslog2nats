package syslog2nats

import "testing"

func TestProduce(t *testing.T) {
	srv := RunBasicJetStreamServer(NATSPORT)
	if srv == nil {
		t.Fatalf("cannot start broker")
	}
	defer shutdownJSServerAndRemoveStorage(t, srv)

	mp := newMsgProducer()

	err := mp.Connect(ConfFact(), NewServerConnection(false))
	if err != nil {
		t.Fatal(err)
	}

	defer mp.Disconnect()
}
