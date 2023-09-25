package syslog2nats

import (
	"testing"
	"time"

	"github.com/g41797/sputnik"
)

func TestProduceConsume(t *testing.T) {
	srv := RunBasicJetStreamServer(NATSPORT)
	if srv == nil {
		t.Fatalf("cannot start broker")
	}
	defer shutdownJSServerAndRemoveStorage(t, srv)

	servconn := NewServerConnection(true)
	defer CloseServerConnection(servconn)

	var mc msgConsumer
	err := mc.Connect(ConfFact(), servconn)
	if err != nil {
		t.Fatal(err)
	}
	defer mc.Disconnect()

	comm := newCommunicator()

	err = mc.Consume(comm)
	if err != nil {
		t.Fatal(err)
	}

	start := comm.Recv()
	if start == nil {
		t.Fatalf("wrong flow")
	}

	var mp msgProducer
	err = mp.Connect(ConfFact(), servconn)
	if err != nil {
		t.Fatal(err)
	}

	defer mp.Disconnect()

	propname := "content"
	propvalue := propname

	pmsg := make(sputnik.Msg)
	pmsg[propname] = propvalue

	err = mp.Produce(pmsg)
	if err != nil {
		t.Fatal(err)
	}

	if !mp.waitAsyncProduce(time.Second) {
		t.Fatalf("timeout of produce")
	}

	state := mc.StreamInfo().State
	if state.Msgs != 1 {
		t.Fatalf("Expected 1 message Actual %d", state.Msgs)
	}

	cmsg := comm.Recv()
	if cmsg == nil {
		t.Fatalf("consume failed")
	}

	actual, exist := cmsg[propname]

	if !exist {
		t.Fatalf("message property does not exist")
	}

	actualtext, ok := actual.(string)
	if !ok {
		t.Fatalf("message property is not text")
	}

	if actualtext != propvalue {
		t.Fatalf("expected %s actual %s", propvalue, actualtext)
	}
}
