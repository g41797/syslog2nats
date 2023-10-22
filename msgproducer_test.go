package syslog2nats

import (
	"testing"
	"time"

	"github.com/g41797/syslogsidecar"
	_ "github.com/g41797/syslogsidecar"
)

func TestProduceConsume(t *testing.T) {
	srv := RunBasicJetStreamServer(NATSPORT)
	if srv == nil {
		t.Fatalf("cannot start broker")
	}
	defer shutdownJSServerAndRemoveStorage(t, srv)

	servconn := NewServerConnection(true)
	defer CloseServerConnection(servconn)

	maxMessages := 1000000

	var mc msgConsumer
	err := mc.Connect(ConfFact(), servconn)
	if err != nil {
		t.Fatal(err)
	}
	defer mc.Disconnect()

	comm := newCommunicator(maxMessages)

	err = mc.Consume(comm)
	if err != nil {
		t.Fatal(err)
	}

	start := comm.Recv(time.Second)
	if start == nil {
		t.Fatalf("wrong flow")
	}

	var mp msgProducer
	merr := mp.Connect(ConfFact(), servconn)
	if merr != nil {
		t.Fatal(merr)
	}
	defer mp.Disconnect()

	go batchProduce(t, &mp, maxMessages)

	batchConsume(t, &mc, comm, maxMessages)

	if !mp.waitAsyncProduce(time.Second) {
		t.Fatalf("timeout of produce")
	}

	state := mc.StreamInfo().State
	if state.Msgs != uint64(maxMessages) {
		t.Fatalf("Expected %d messages Actual %d", maxMessages, state.Msgs)
	}
}

func batchProduce(t *testing.T, mp *msgProducer, count int) {
	badmsg := map[string]string{
		syslogsidecar.Formermessage: syslogsidecar.Formermessage}

	for i := 0; i < count; i++ {
		pmsg := syslogsidecar.Get()

		err := syslogsidecar.Pack(pmsg, badmsg)
		if err != nil {
			t.Error(err)
			return
		}

		err = mp.Produce(pmsg)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func batchConsume(t *testing.T, mc *msgConsumer, comm *dumbCommunicator, count int) {
	propname := syslogsidecar.Formermessage
	propvalue := propname

	for i := 0; i < count; i++ {
		cmsg := comm.Recv(100 * time.Millisecond)
		if cmsg == nil {
			t.Fatalf("consume failed")
		}

		consumed, exist := cmsg["consumed"]

		if !exist {
			t.Fatalf("message property consumed does not exist")
		}

		consmap, _ := consumed.(map[string]string)

		actualtext, ok := consmap[propname]

		if !ok {
			t.Fatalf("message property is not text")
		}

		if actualtext != propvalue {
			t.Fatalf("expected %s actual %s", propvalue, actualtext)
		}
	}
}
