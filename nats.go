package syslog2nats

import (
	"os"

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

func ShutdownJSServerAndRemoveStorage(s *server.Server) {
	var sd string
	if config := s.JetStreamConfig(); config != nil {
		sd = config.StoreDir
	}
	s.Shutdown()
	if sd != "" {
		os.RemoveAll(sd)
	}
	s.WaitForShutdown()
}
