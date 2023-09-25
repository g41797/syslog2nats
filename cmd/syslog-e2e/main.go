package main

import (
	"github.com/g41797/sputnik/sidecar"
	"github.com/g41797/syslog2nats"

	// Attach blocks and plugins to the process:
	_ "github.com/g41797/sputnik"
	_ "github.com/g41797/syslog2nats"
	_ "github.com/g41797/syslogsidecar"
	_ "github.com/g41797/syslogsidecar/e2e"
)

func main() {

	srv := syslog2nats.RunBasicJetStreamServer(syslog2nats.NATSPORT)
	if srv == nil {
		panic("cannot start broker")
	}
	defer syslog2nats.ShutdownJSServerAndRemoveStorage(srv)

	sidecar.Start(syslog2nats.NewConnector())
}
