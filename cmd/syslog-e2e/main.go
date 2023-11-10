package main

import (
	"embed"

	"github.com/g41797/sputnik/sidecar"
	"github.com/g41797/syslog2nats"

	// Attach blocks and plugins to the process:
	_ "github.com/g41797/sputnik"
	_ "github.com/g41797/syslog2nats"
	_ "github.com/g41797/syslogsidecar"
	_ "github.com/g41797/syslogsidecar/e2e"
)

//go:embed conf
var embconf embed.FS

func main() {

	srv := syslog2nats.RunBasicJetStreamServer(syslog2nats.NATSPORT)
	if srv == nil {
		panic("cannot start broker")
	}
	defer syslog2nats.ShutdownJSServerAndRemoveStorage(srv)

	cleanUp, _ := sidecar.UseEmbeddedConfiguration(&embconf)
	defer cleanUp()
	sidecar.Start(syslog2nats.NewConnector())
}
