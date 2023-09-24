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

type msgProducer struct {
	conf MsgPrdConfig
	sc   *natsConnection
	js   jetstream.JetStream
	ctx  context.Context
}

func (mpr *msgProducer) Connect(cf sputnik.ConfFactory, scn sputnik.ServerConnection) error {
	err := cf(MsgProducerConfigName, &mpr.conf)
	if err != nil {
		return err
	}

	sc := scn.(*natsConnection)

	js, err := jetstream.New(sc.nc)
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

	if mpr.sc.shared {
		mpr.sc = nil
		return
	}

	mpr.sc.nc.Close()
	mpr.sc = nil
	return
}

func (mpr *msgProducer) Produce(msg sputnik.Msg) error {

	if mpr.sc == nil {
		return fmt.Errorf("connection with broker does not exist")
	}

	if mpr.sc.nc == nil {
		return fmt.Errorf("connection with broker does not exist")
	}

	if !mpr.sc.nc.IsConnected() {
		return fmt.Errorf("does not connected with broker")
	}

	natsmsg := mpr.convertProduceMsg(msg)

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

func (mpr *msgProducer) convertProduceMsg(inmsg sputnik.Msg) *nats.Msg {
	msg := &nats.Msg{
		Subject: mpr.conf.STREAM,
		Header:  make(nats.Header),
	}

	if inmsg == nil {
		return msg
	}

	if len(inmsg) == 0 {
		return msg
	}

	for k, v := range inmsg {
		vstr, ok := v.(string)
		if !ok {
			continue
		}
		msg.Header.Add(k, vstr)
	}

	return msg
}
