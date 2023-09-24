package syslog2nats

import (
	"testing"
	"time"

	"github.com/g41797/sputnik"
)

func TestProduce(t *testing.T) {
	srv := RunBasicJetStreamServer(NATSPORT)
	if srv == nil {
		t.Fatalf("cannot start broker")
	}
	defer shutdownJSServerAndRemoveStorage(t, srv)

	var mp msgProducer

	err := mp.Connect(ConfFact(), NewServerConnection(false))
	if err != nil {
		t.Fatal(err)
	}

	pmsg := make(sputnik.Msg)
	pmsg["content"] = "content"

	err = mp.Produce(pmsg)
	if err != nil {
		t.Fatal(err)
	}

	if !mp.waitAsyncProduce(time.Second) {
		t.Fatalf("timeout of produce")
	}

	defer mp.Disconnect()
}
