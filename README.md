# syslog2nats [![Go](https://github.com/g41797/syslog2nats/actions/workflows/go.yml/badge.svg)](https://github.com/g41797/syslog2nats/actions/workflows/go.yml)

Receives syslog messages and publishs them to [NATS](https://nats.io) 

syslog2nats is based on 
- [syslogsidecar](https://github.com/g41797/syslogsidecar#readme)
- [sputnik](https://github.com/g41797/sputnik)

syslog2nats consists of:
- syslog server - common part for all syslogsidecar based processes
- NATS specific plugins 

## Syslog server

 Supported RFCs:
  - [RFC3164](<https://tools.ietf.org/html/rfc3164>)
  - [RFC5424](<https://tools.ietf.org/html/rfc5424>)


  RFC3164 message consists of following symbolic parts:
  - priority
  - facility 
  - severity
  - timestamp
  - hostname
  - tag
  - **content**

  ### RFC5424

  RFC5424 message consists of following symbolic parts:
 - priority
 - facility 
 - severity
 - timestamp
 - hostname
 - version
 - app_name
 - proc_id
 - msg_id
 - structured_data
 - **message**

  ### Non-RFC parts

  syslogsidecar adds following parts to standard ones:
  - rfc of produced message:
    - Part name: "rfc"
    - Values: "RFC3164"|"RFC5424"
  - former syslog message:
    - Part name: "data"
  - parsing error (optional):
    - Part name: "parsererror"

      
### Severities

    Valid severity levels and names are:

 - 0 emerg
 - 1 alert
 - 2 crit
 - 3 err
 - 4 warning
 - 5 notice
 - 6 info
 - 7 debug

  syslogsidecar filters messages by level according to value in configuration, e.g. for:
```json
{
  "SEVERITYLEVEL": 4,
  ...........
}
```
all messages with severity above 4 will be discarded. 


  ### Configuration

  Configuration of syslog server part of syslogsidecar is saved in the file syslogreceiver.json:
```json
{
    "SEVERITYLEVEL": 4,
    "ADDRTCP": "127.0.0.1:5141",
    "ADDRUDP": "127.0.0.1:5141",
    "UDSPATH": "",
    "ADDRTCPTLS": "127.0.0.1:5143",
    "CLIENT_CERT_PATH": "",
    "CLIENT_KEY_PATH ": "",
    "ROOT_CA_PATH": ""
}
```

### Links

- More complete description of [syslogsidecar](https://github.com/g41797/syslogsidecar#readme)
- syslog for [Memphis](https://memphis.dev) is part of [memphis-protocol-adapter](https://github.com/g41797/memphis-protocol-adapter) project


## Plugins

  NATS plugins to syslogsidecar:
  - [connector](https://github.com/g41797/syslog2nats/blob/main/connector.go)
  - [producer](https://github.com/g41797/syslog2nats/blob/main/msgproducer.go)
  - [consumer](https://github.com/g41797/syslog2nats/blob/main/msgconsumer.go) (used for the tests)


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

More about connector and underlying software - [sputnik](https://github.com/g41797/sputnik#readme)

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

## Build and run under vscode

```bash
go clean -cache -testcache
go build ./cmd/syslog2nats/
./syslog2nats -cf ./cmd/syslog2nats/conf/
```

## e2e test

Simultaneuosly:  
- send 1000000 syslog messages
   - TCP/IP
   - RFC5424
- receive
- produce
- consume
- compare

Build and run under vscode:
```bash
go clean -cache -testcache
go build ./cmd/syslog-e2e/
./syslog-e2e -cf ./cmd/syslog-e2e/conf/
```
nats server runs as embedded within syslog-e2e process.

Report:
20.59306566s   Was send 1000000 messages. Successfully consumed 1000000 Received 1000000

