package main

import (
	"github.com/g41797/sputnik/sidecar"
	"github.com/g41797/syslog2nats"

	// Attach blocks and plugins to the process:
	_ "github.com/g41797/sputnik"
	_ "github.com/g41797/syslogsidecar"
)

func main() {
	sidecar.Start(syslog2nats.NewConnector())
}
