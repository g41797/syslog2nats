package syslog2nats

import (
	"context"
	"time"

	"github.com/g41797/sputnik"
	"github.com/g41797/sputnik/sidecar"
	"github.com/g41797/syslogsidecar"
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
	/*
		if mpr.mc == nil {
			return fmt.Errorf("connection with broker does not exist")
		}

		if !mpr.mc.IsConnected() {
			return fmt.Errorf("does not connected with broker")
		}

		hdrs := memphis.Headers{}
		hdrs.New()

		for k, v := range msg {
			vstr, ok := v.(string)
			if !ok {
				continue
			}
			if err := hdrs.Add(k, vstr); err != nil {
				return err
			}
		}

		err := mpr.producer.Produce("", memphis.MsgHeaders(hdrs))

		return err
	*/
	return nil
}

func (mpr *msgProducer) waitAsyncProduce(to time.Duration) {
	select {
	case <-mpr.js.PublishAsyncComplete():
	case <-time.After(to):
		return
	}
}
