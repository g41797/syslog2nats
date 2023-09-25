package syslog2nats

import (
	"os"
	"testing"
	"time"

	"github.com/g41797/sputnik"
	"github.com/g41797/sputnik/sidecar"
	"github.com/nats-io/nats-server/v2/server"

	natsserver "github.com/nats-io/nats-server/v2/test"
)

const (
	NATSPORT = 4222
	CONFPATH = "./conf"
)

func RunBasicJetStreamServer(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	opts.JetStream = true
	return RunServerWithOptions(opts)
}

func RunServerWithOptions(opts server.Options) *server.Server {
	return natsserver.RunServer(&opts)
}

func shutdownJSServerAndRemoveStorage(t *testing.T, s *server.Server) {
	t.Helper()
	var sd string
	if config := s.JetStreamConfig(); config != nil {
		sd = config.StoreDir
	}
	s.Shutdown()
	if sd != "" {
		if err := os.RemoveAll(sd); err != nil {
			t.Fatalf("Unable to remove storage %q: %v", sd, err)
		}
	}
	s.WaitForShutdown()
}

func ConfFact() sputnik.ConfFactory {
	return sidecar.ConfigFactory(CONFPATH)
}

// connector always returns Shared connection
// for usage this connection with tests we need more flexible approach
func NewServerConnection(Shared bool) sputnik.ServerConnection {
	cntr := newConnector()

	scn, err := cntr.Connect(ConfFact())

	if err != nil {
		return nil
	}

	nonshared := scn.(*NatsConnection)
	nonshared.Shared = Shared

	return nonshared
}

func CloseServerConnection(sconn sputnik.ServerConnection) {
	if sconn == nil {
		return
	}
	sc := sconn.(*NatsConnection)

	if sc.NatsConn != nil {
		sc.NatsConn.Close()
	}
	return
}

var _ sputnik.BlockCommunicator = &dumbCommunicator{}

type dumbCommunicator struct {
	msgs chan sputnik.Msg
}

func newCommunicator() *dumbCommunicator {
	return &dumbCommunicator{msgs: make(chan sputnik.Msg, 3)}
}

func (c *dumbCommunicator) Communicator(resp string) (bc sputnik.BlockCommunicator, exists bool) {
	return nil, false
}

func (c *dumbCommunicator) Descriptor() sputnik.BlockDescriptor {
	return sputnik.BlockDescriptor{}
}

func (c *dumbCommunicator) Send(msg sputnik.Msg) bool {
	c.msgs <- msg
	return true
}

func (c *dumbCommunicator) Recv() sputnik.Msg {
	select {
	case msg := <-c.msgs:
		return msg
	case <-time.After(10 * time.Second):
		return nil
	}
}
