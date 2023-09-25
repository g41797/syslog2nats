package syslog2nats

import (
	"context"

	"github.com/g41797/sputnik"
	"github.com/g41797/sputnik/sidecar"
	"github.com/g41797/syslogsidecar/e2e"
	"github.com/nats-io/nats.go/jetstream"
)

func init() {
	e2e.RegisterMessageConsumerFactory(newMsgConsumer)
}

const MsgConsumerConfigName = MsgProducerConfigName

type msgConsumer struct {
	conf    MsgPrdConfig
	sc      *NatsConnection
	js      jetstream.JetStream
	stream  jetstream.Stream
	cons    jetstream.Consumer
	stop    jetstream.ConsumeContext
	ctx     context.Context
	sender  sputnik.BlockCommunicator
	started bool
}

func newMsgConsumer() sidecar.MessageConsumer {
	return new(msgConsumer)
}

func (mcn *msgConsumer) Connect(cf sputnik.ConfFactory, Shared sputnik.ServerConnection) error {
	err := cf(MsgConsumerConfigName, &mcn.conf)
	if err != nil {
		return err
	}

	sc := Shared.(*NatsConnection)

	js, err := jetstream.New(sc.NatsConn)
	if err != nil {
		return err
	}

	ctx := context.Background()
	mcn.stream, err = js.CreateStream(ctx, jetstream.StreamConfig{Name: mcn.conf.STREAM})
	if err != nil {
		return err
	}

	mcn.stream.Purge(ctx)

	mcn.cons, err = js.CreateOrUpdateConsumer(ctx, mcn.conf.STREAM, jetstream.ConsumerConfig{
		HeadersOnly: true,
	})

	if err != nil {
		return err
	}

	mcn.sc = sc
	mcn.js = js
	mcn.ctx = ctx

	return nil

}

func (cons *msgConsumer) Consume(sender sputnik.BlockCommunicator) error {
	if cons.started {
		return nil
	}

	cons.sender = sender

	var err error

	cons.stop, err = cons.cons.Consume(cons.onMessage)
	if err != nil {
		cons.stop = nil
		return err
	}

	cons.startTest()
	cons.started = true
	return nil
}

func (cons *msgConsumer) onMessage(inmsg jetstream.Msg) {
	inmsg.Ack()
	if cons.sender == nil {
		return
	}
	omsg := ConvertConsumeMsg(inmsg)

	if omsg == nil {
		return
	}
	cons.sender.Send(omsg)
	return
}

func (cons *msgConsumer) Disconnect() {
	if cons == nil {
		return
	}

	if cons.stop != nil {
		cons.stop.Stop()
	}

	if (cons.sc != nil) && (!cons.sc.Shared) {
		cons.sc.NatsConn.Close()
	}

	cons.stopTest()
	cons.started = false
	return
}

func ConvertConsumeMsg(inmsg jetstream.Msg) sputnik.Msg {
	if inmsg == nil {
		return nil
	}

	headers := inmsg.Headers()

	if headers == nil {
		return nil
	}

	if len(headers) == 0 {
		return nil
	}

	props := make(map[string]string)

	for k, v := range headers {
		props[k] = v[0]
	}

	smsg := sputnik.Msg{}
	smsg["name"] = "consumed"
	smsg["consumed"] = props
	smsg["data"] = ""

	return smsg
}

func (cons *msgConsumer) startTest() {
	if cons.sender == nil {
		return
	}
	msg := sputnik.Msg{}
	msg["name"] = "start"
	cons.sender.Send(msg)
}

func (cons *msgConsumer) stopTest() {
	if cons.sender == nil {
		return
	}
	msg := sputnik.Msg{}
	msg["name"] = "stop"
	cons.sender.Send(msg)
}

func (cons *msgConsumer) StreamInfo() *jetstream.StreamInfo {
	if cons.stream == nil {
		return &jetstream.StreamInfo{}
	}

	streamInfo, err := cons.stream.Info(cons.ctx)

	if err != nil {
		return &jetstream.StreamInfo{}
	}

	return streamInfo
}
