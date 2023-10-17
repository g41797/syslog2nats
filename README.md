# syslog2nats [![Go](https://github.com/g41797/syslog2nats/actions/workflows/go.yml/badge.svg)](https://github.com/g41797/syslog2nats/actions/workflows/go.yml)

Example of using [syslogsidecar](https://github.com/g41797/syslogsidecar#readme) with [NATS](https://nats.io) 

## Plugins

  In order to supply NATS specific functionality, 3 plugins to syslogsidecar were developed:
  - connector
  - producer
  - consumer (used for the tests)


### Connector

Configuration file: connector.json
```json
{
    "CONN_URL": "nats://127.0.0.1:4222"
}
```
The rest of connection options are default.

Connector creates sharable _*nats.Conn*_ for:
- periodic validation of connectivity with memphis
- using by producer (production) and consumer (e2e test)

### Producer

Configuration file: syslogproducer.json
```json
{
    "STREAM": "syslog"
}
```

Producer uses received from connector _*nats.Conn*_.
It created JETSTREAM with name from configuration, the rest of stream options are default.

syslog messages are produced to jetstream as *Header* with _*empty payload*_:
```go
    .................................
    msg := &nats.Msg{
		Subject: name,
		Header:  make(nats.Header),
	}

	for k, v := range inmsg {
		vstr, ok := v.(string)
		if !ok {
			continue
		}
		msg.Header.Add(k, vstr)
	}
    .................................
```

## e2e test

- TCP/IP 
- 1000000 syslog messages 
- received
- produced(published)
- consumed
- compared

Build and run under vscode:
```bash
go clean -cache -testcache
go build ./cmd/syslog-e2e/
./syslog-e2e -cf ./cmd/syslog-e2e/conf/
```
nats server runs as embedded within syslog-e2e process.

Report:
20.59306566s   Was send 1000000 messages. Successfully consumed 1000000 Received 1000000

