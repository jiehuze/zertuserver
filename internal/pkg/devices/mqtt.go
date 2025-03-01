package devices

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var (
	// ErrNilClient is returned when a client action can't be taken because the struct has no client
	ErrNilClient = errors.New("no MQTT client available")
)

// Message is a message received from the broker.
type Message paho.Message

// Adaptor is the Gobot Adaptor for MQTT
type MqttDevice struct {
	name          string
	Host          string
	clientID      string
	username      string
	password      string
	useSSL        bool
	serverCert    string
	clientCert    string
	clientKey     string
	autoReconnect bool
	cleanSession  bool
	client        paho.Client
	qos           int
}

// NewAdaptor creates a new mqtt adaptor with specified host and client id
func NewMqttDevice(host string, clientID string) *MqttDevice {
	return &MqttDevice{
		name:          "MQTT",
		Host:          host,
		autoReconnect: false,
		cleanSession:  true,
		useSSL:        false,
		clientID:      clientID,
	}
}

// NewAdaptorWithAuth creates a new mqtt adaptor with specified host, client id, username, and password.
func NewMqttDeviceWithAuth(host, clientID, username, password string) *MqttDevice {
	return &MqttDevice{
		name:          "MQTT",
		Host:          host,
		autoReconnect: false,
		cleanSession:  true,
		useSSL:        false,
		clientID:      clientID,
		username:      username,
		password:      password,
	}
}

// Name returns the MQTT Adaptor's name
func (a *MqttDevice) Name() string { return a.name }

// SetName sets the MQTT Adaptor's name
func (a *MqttDevice) SetName(n string) { a.name = n }

// Port returns the Host name
func (a *MqttDevice) Port() string { return a.Host }

// AutoReconnect returns the MQTT AutoReconnect setting
func (a *MqttDevice) AutoReconnect() bool { return a.autoReconnect }

// SetAutoReconnect sets the MQTT AutoReconnect setting
func (a *MqttDevice) SetAutoReconnect(val bool) { a.autoReconnect = val }

// CleanSession returns the MQTT CleanSession setting
func (a *MqttDevice) CleanSession() bool { return a.cleanSession }

// SetCleanSession sets the MQTT CleanSession setting. Should be false if reconnect is enabled. Otherwise all subscriptions will be lost
func (a *MqttDevice) SetCleanSession(val bool) { a.cleanSession = val }

// UseSSL returns the MQTT server SSL preference
func (a *MqttDevice) UseSSL() bool { return a.useSSL }

// SetUseSSL sets the MQTT server SSL preference
func (a *MqttDevice) SetUseSSL(val bool) { a.useSSL = val }

// ServerCert returns the MQTT server SSL cert file
func (a *MqttDevice) ServerCert() string { return a.serverCert }

// SetQoS sets the QoS value passed into the MTT client on Publish/Subscribe events
func (a *MqttDevice) SetQoS(qos int) { a.qos = qos }

// SetServerCert sets the MQTT server SSL cert file
func (a *MqttDevice) SetServerCert(val string) { a.serverCert = val }

// ClientCert returns the MQTT client SSL cert file
func (a *MqttDevice) ClientCert() string { return a.clientCert }

// SetClientCert sets the MQTT client SSL cert file
func (a *MqttDevice) SetClientCert(val string) { a.clientCert = val }

// ClientKey returns the MQTT client SSL key file
func (a *MqttDevice) ClientKey() string { return a.clientKey }

// SetClientKey sets the MQTT client SSL key file
func (a *MqttDevice) SetClientKey(val string) { a.clientKey = val }

// Connect returns true if connection to mqtt is established
func (a *MqttDevice) Connect() (err error) {
	a.client = paho.NewClient(a.createClientOptions())
	if token := a.client.Connect(); token.Wait() && token.Error() != nil {
		//err = multierror.Append(err, token.Error())
		//return token.Error()
		return errors.New("connect error!")
	}

	return
}

// Disconnect returns true if connection to mqtt is closed
func (a *MqttDevice) Disconnect() (err error) {
	if a.client != nil {
		a.client.Disconnect(500)
	}
	return
}

// Finalize returns true if connection to mqtt is finalized successfully
func (a *MqttDevice) Finalize() (err error) {
	a.Disconnect()
	return
}

// Publish a message under a specific topic
func (a *MqttDevice) Publish(topic string, message []byte) bool {
	_, err := a.PublishWithQOS(topic, a.qos, message)
	if err != nil {
		return false
	}

	return true
}

// PublishAndRetain publishes a message under a specific topic with retain flag
func (a *MqttDevice) PublishAndRetain(topic string, message []byte) bool {
	if a.client == nil {
		return false
	}

	a.client.Publish(topic, byte(a.qos), true, message)
	return true
}

// PublishWithQOS allows per-publish QOS values to be set and returns a paho.Token
func (a *MqttDevice) PublishWithQOS(topic string, qos int, message []byte) (paho.Token, error) {
	if a.client == nil {
		return nil, ErrNilClient
	}

	token := a.client.Publish(topic, byte(qos), false, message)
	return token, nil
}

// OnWithQOS allows per-subscribe QOS values to be set and returns a paho.Token
func (a *MqttDevice) Subscribe(topic string, qos int, f func(msg Message)) (paho.Token, error) {
	if a.client == nil {
		return nil, ErrNilClient
	}

	token := a.client.Subscribe(topic, byte(qos), func(client paho.Client, msg paho.Message) {
		f(msg)
	})

	return token, nil
}

func (a *MqttDevice) createClientOptions() *paho.ClientOptions {
	opts := paho.NewClientOptions()
	opts.AddBroker(a.Host)
	opts.SetClientID(a.clientID)
	if a.username != "" && a.password != "" {
		opts.SetPassword(a.password)
		opts.SetUsername(a.username)
	}
	opts.AutoReconnect = a.autoReconnect
	opts.CleanSession = a.cleanSession

	if a.UseSSL() {
		opts.SetTLSConfig(a.newTLSConfig())
	}
	return opts
}

// newTLSConfig sets the TLS config in the case that we are using
// an MQTT broker with TLS
func (a *MqttDevice) newTLSConfig() *tls.Config {
	// Import server certificate
	var certpool *x509.CertPool
	if len(a.ServerCert()) > 0 {
		certpool = x509.NewCertPool()
		pemCerts, err := ioutil.ReadFile(a.ServerCert())
		if err == nil {
			certpool.AppendCertsFromPEM(pemCerts)
		}
	}

	// Import client certificate/key pair
	var certs []tls.Certificate
	if len(a.ClientCert()) > 0 && len(a.ClientKey()) > 0 {
		cert, err := tls.LoadX509KeyPair(a.ClientCert(), a.ClientKey())
		if err != nil {
			// TODO: proper error handling
			panic(err)
		}
		certs = append(certs, cert)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		Certificates: certs,
	}
}
