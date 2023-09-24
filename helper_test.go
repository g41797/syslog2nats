package syslog2nats

import (
	"os"
	"testing"

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

func NewServerConnection() sputnik.ServerConnection {
	cntr := newConnector()

	scn, err := cntr.Connect(ConfFact())

	if err != nil {
		return nil
	}

	nonshared := scn.(natsConnection)
	nonshared.shared = false

	return nonshared
}
