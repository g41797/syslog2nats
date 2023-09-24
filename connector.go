package syslog2nats

import (
	"github.com/g41797/sputnik"
	"github.com/nats-io/nats.go"
	_ "github.com/nats-io/nats.go/jetstream"
)

const connectorConfName = "connector"

type brokerConnConfig struct {
	CONN_URL string
}

var _ sputnik.ServerConnector = &natsConnector{}

func newConnector() sputnik.ServerConnector {
	return new(natsConnector)
}

type natsConnection struct {
	shared bool
	nc     *nats.Conn
}

type natsConnector struct {
	sc *natsConnection
}

func (c *natsConnector) Connect(cf sputnik.ConfFactory) (conn sputnik.ServerConnection, err error) {
	if c.IsConnected() {
		return c.sc, nil
	}

	var conf brokerConnConfig

	if err = cf(connectorConfName, &conf); err != nil {
		return nil, err
	}

	nc, err := nats.Connect(conf.CONN_URL)
	if err != nil {
		return nil, err
	}

	c.sc = &natsConnection{true, nc}

	return c.sc, nil
}

func (c *natsConnector) IsConnected() bool {
	if c.sc == nil {
		return false
	}

	if c.sc.nc.IsClosed() {
		return false
	}

	return true
}

func (c *natsConnector) Disconnect() {
	if c.sc.nc != nil {
		c.sc.nc.Close()
	}
	c.sc = nil
}

/*
const (
	sourceName         = "prtcl-adptr"
	syslogsStreamName  = "$memphis_syslogs"
	syslogsInfoSubject = "extern.info"
	syslogsWarnSubject = "extern.warn"
	syslogsErrSubject  = "extern.err"
)

type BrokerConnConfig struct {
	MEMPHIS_ADDR         string
	MEMPHIS_CLIENT       string
	ROOT_USER            string
	ROOT_PASSWORD        string
	CONNECTION_TOKEN     string
	CLIENT_CERT_PATH     string
	CLIENT_KEY_PATH      string
	ROOT_CA_PATH         string
	USER_PASS_BASED_AUTH bool
	DEBUG                bool
	CLOUD_ENV            bool
	DEV_ENV              bool
}

var _ sputnik.ServerConnector = &BrokerConnector{}
var _ io.Writer = &BrokerConnector{}
var _ LoggerFactory = new(BrokerConnector).getLogger

const connectorConfName = "connector"

type BrokerConnector struct {
	io.Writer
	conf       BrokerConnConfig
	nc         *nats.Conn
	l          atomic.Pointer[Logger]
	flags      int
	pidPrefix  string
	labelStart int
	baseSubj   string
	lblToSubj  map[string]string
	tlsConfig  *tls.Config
}

func (c *BrokerConnector) Connect(cf sputnik.ConfFactory) (conn sputnik.ServerConnection, err error) {
	if c.IsConnected() {
		return c.getLogger, nil
	}

	var conf BrokerConnConfig

	if err = cf(connectorConfName, &conf); err != nil {
		return nil, err
	}

	return c.ConnectWithConfig(conf)
}

func (c *BrokerConnector) ConnectWithConfig(conf BrokerConnConfig) (conn sputnik.ServerConnection, err error) {

	c.conf = conf

	if err = c.prepareTLS(); err != nil {
		return nil, err
	}

	if err = c.connect(); err != nil {
		return nil, err
	}

	c.createLogger()

	return c.getLogger, nil
}

// For first implementation advanced callbacks of nats.Conn
// (connected/disconnected/reconnected) are not used
// Implementation from rest-gateway
func (c *BrokerConnector) connect() error {
	var nc *nats.Conn
	var err error

	natsOpts := nats.Options{
		Url:            c.conf.MEMPHIS_ADDR,
		AllowReconnect: true,
		MaxReconnect:   10,
		ReconnectWait:  3 * time.Second,
		Name:           c.conf.MEMPHIS_CLIENT,
	}

	creds := c.conf.CONNECTION_TOKEN
	username := c.conf.ROOT_USER
	if c.conf.USER_PASS_BASED_AUTH {
		username = "$$memphis"
		natsOpts.User = username
		creds = c.conf.CONNECTION_TOKEN + "_" + c.conf.ROOT_PASSWORD
		natsOpts.Password = creds
	} else {
		natsOpts.Token = username + "::" + creds
	}

	if c.tlsConfig != nil {
		natsOpts.TLSConfig = c.tlsConfig
	}

	nc, err = natsOpts.Connect()
	if err != nil {
		return err
	}

	c.nc = nc

	return nil
}

func (c *BrokerConnector) prepareTLS() error {

	if c.tlsConfig != nil {
		return nil
	}

	t, err := PrepareTLS(c.conf.CLIENT_CERT_PATH, c.conf.CLIENT_KEY_PATH, c.conf.ROOT_CA_PATH)

	if err != nil {
		return err
	}

	c.tlsConfig = t

	return nil
}

func (c *BrokerConnector) IsConnected() bool {
	if c == nil {
		return false
	}

	if c.nc == nil {
		return false
	}

	if !c.nc.IsConnected() {
		if !c.nc.IsClosed() {
			c.nc.Close()
		}
		c.l.Store(nil)
		return false
	}

	return true
}

func (c *BrokerConnector) Disconnect() {
	if !c.IsConnected() {
		return
	}
	c.l.Store(nil)
	c.nc.Close()
	return
}

func (c *BrokerConnector) getLogger() *Logger {
	if c == nil {
		return nil
	}
	return c.l.Load()
}

func (c *BrokerConnector) createLogger() {
	// From rest-gateway implementation
	c.flags = log.LstdFlags | log.Lmicroseconds
	c.pidPrefix = fmt.Sprintf("[%d] ", os.Getpid())
	c.labelStart = len(c.pidPrefix) + 28 //???
	c.baseSubj = fmt.Sprintf("%s.%s.", syslogsStreamName, sourceName)
	c.lblToSubj = map[string]string{
		"INF": syslogsInfoSubject,
		"WRN": syslogsWarnSubject,
		"ERR": syslogsErrSubject,
	}

	c.l.Store(NewLogger(log.New(c, c.pidPrefix, c.flags)))

	return
}

func (c *BrokerConnector) Write(p []byte) (n int, err error) {
	if !c.IsConnected() {
		return 0, errors.New("not connected")
	}

	if c.conf.CLOUD_ENV {
		return len(p), nil
	}

	label := string(p[c.labelStart : c.labelStart+labelLen])
	subjectSuffix, ok := c.lblToSubj[label]
	if !ok { // skip other labels
		return 0, nil
	}

	subject := c.baseSubj + subjectSuffix

	if err := c.nc.Publish(subject, p); err != nil {
		return 0, err
	}

	return len(p), nil
}

func PrepareTLS(CLIENT_CERT_PATH, CLIENT_KEY_PATH, ROOT_CA_PATH string) (*tls.Config, error) {

	if CLIENT_CERT_PATH == "" || CLIENT_KEY_PATH != "" || ROOT_CA_PATH != "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(CLIENT_CERT_PATH, CLIENT_KEY_PATH)
	if err != nil {
		return nil, err
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	TLSConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	TLSConfig.Certificates = []tls.Certificate{cert}
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(ROOT_CA_PATH)
	if err != nil {
		return nil, err
	}
	certs.AppendCertsFromPEM(pemData)
	TLSConfig.RootCAs = certs

	return TLSConfig, nil
}
*/
