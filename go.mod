module github.com/g41797/syslog2nats

go 1.20

require (
	github.com/g41797/sputnik v0.0.17
	github.com/g41797/syslogsidecar v0.0.12
	github.com/nats-io/nats-server/v2 v2.10.1
	github.com/nats-io/nats.go v1.30.0
)

require (
	github.com/RackSec/srslog v0.0.0-20180709174129-a4725f04ec91 // indirect
	github.com/RoaringBitmap/roaring v1.5.0 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/g41797/go-syslog v1.0.5 // indirect
	github.com/g41797/gonfig v1.0.1 // indirect
	github.com/g41797/kissngoqueue v0.1.5 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/jwt/v2 v2.5.2 // indirect
	github.com/nats-io/nkeys v0.4.5 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/g41797/syslogsidecar v0.0.10 => ../syslogsidecar
