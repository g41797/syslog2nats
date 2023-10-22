package syslog2nats

import (
	"context"
	"fmt"
	"time"

	"github.com/g41797/sputnik"
	"github.com/g41797/sputnik/sidecar"
	"github.com/g41797/syslogsidecar"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func init() {
	syslogsidecar.RegisterMessageProducerFactory(newMsgProducer)
}

const MsgProducerConfigName = syslogsidecar.ProducerName

type MsgPrdConfig struct {
	STREAM string
}

func newMsgProducer() sidecar.MessageProducer {
	return &msgProducer{}
}

var _ sidecar.MessageProducer = &msgProducer{}

type msgProducer struct {
	conf MsgPrdConfig
	sc   *NatsConnection
	js   jetstream.JetStream
	ctx  context.Context
}

func (mpr *msgProducer) Connect(cf sputnik.ConfFactory, scn sputnik.ServerConnection) error {
	err := cf(MsgProducerConfigName, &mpr.conf)
	if err != nil {
		return err
	}

	sc := scn.(*NatsConnection)

	js, err := jetstream.New(sc.NatsConn)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = js.CreateStream(ctx, jetstream.StreamConfig{Name: mpr.conf.STREAM})
	if err != nil {
		return err
	}
	mpr.sc = sc
	mpr.js = js
	mpr.ctx = ctx
	return nil
}

func (mpr *msgProducer) Disconnect() {
	if mpr.sc == nil {
		return
	}

	if mpr.js != nil {
		mpr.waitAsyncProduce(time.Second)
	}

	if mpr.sc.Shared {
		mpr.sc = nil
		return
	}

	mpr.sc.NatsConn.Close()
	mpr.sc = nil
	return
}

func (mpr *msgProducer) Produce(msg sputnik.Msg) error {

	defer syslogsidecar.Put(msg)

	if mpr.sc == nil {
		return fmt.Errorf("connection with broker does not exist")
	}

	if mpr.sc.NatsConn == nil {
		return fmt.Errorf("connection with broker does not exist")
	}

	if !mpr.sc.NatsConn.IsConnected() {
		return fmt.Errorf("does not connected with broker")
	}

	natsmsg := ConvertProduceMsg(mpr.conf.STREAM, msg)

	if natsmsg == nil {
		return nil
	}

	_, err := mpr.js.PublishMsgAsync(natsmsg)

	return err

}

func (mpr *msgProducer) waitAsyncProduce(to time.Duration) bool {
	select {
	case <-mpr.js.PublishAsyncComplete():
		return true
	case <-time.After(to):
		return false
	}
}

func ConvertProduceMsg(name string, inmsg sputnik.Msg) *nats.Msg {
	msg := &nats.Msg{
		Subject: name,
		Header:  make(nats.Header),
	}

	if inmsg == nil {
		return msg
	}

	if len(inmsg) == 0 {
		return msg
	}

	putToheader := func(name string, value string) error {
		msg.Header.Add(name, value)
		return nil
	}

	syslogsidecar.Unpack(inmsg, putToheader)

	return msg
}
