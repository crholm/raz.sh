package mx

import (
	"bytes"
	"context"
	"github.com/emersion/go-msgauth/dkim"
	"github.com/phires/go-guerrilla"
	"github.com/phires/go-guerrilla/backends"
	"github.com/phires/go-guerrilla/mail"
	"strings"
)

type Config struct {
	Hostname      string `cli:"hostname"`
	SmtpInterface string `cli:"smtp-interface"`

	PublishAddress string `cli:"smtp-publish-address"`
}

type MX struct {
	backend *backend
	daemon  guerrilla.Daemon
	cfg     Config
}

type backend struct {
	mx *MX
}

func New(cfg Config) (*MX, error) {

	cfg.PublishAddress = strings.ToLower(cfg.PublishAddress)

	mx := &MX{
		cfg: cfg,
	}

	mx.backend = &backend{mx: mx}
	mx.daemon.Backend = mx.backend

	mx.daemon.Config = &guerrilla.AppConfig{
		Servers: []guerrilla.ServerConfig{
			{
				Hostname:        cfg.Hostname,
				ListenInterface: cfg.SmtpInterface,

				//TLS: guerrilla.ServerTLSConfig{
				//	StartTLSOn: true,
				//	PrivateKeyFile: "certs/server.key",
				//	PublicKeyFile: "certs/server.crt",
				//},
			},
		},
	}

	return mx, nil
}

func (m *MX) Start() error {

	return m.daemon.Start()
}

func (m *MX) Stop(ctx context.Context) error {
	m.daemon.Shutdown()

	return nil
}

func (m *backend) Start() error {
	return nil
}

func (m *backend) Process(e *mail.Envelope) backends.Result {
	err := m.ValidateRcpt(e)
	if err != nil {
		return backends.NewResult(551, "User not local")
	}

	// TODO: DKIM verification
	// TODO: SPF verification

	title := e.Subject
	data := e.Data.Bytes() // should be header and body

	v, err := dkim.Verify(bytes.NewBuffer(data))

	return backends.NewResult(250, "OK")
}

func (m *backend) ValidateRcpt(e *mail.Envelope) backends.RcptError {
	if len(e.RcptTo) != 1 {
		return backends.NoSuchUser // we only accept 1 recipient
	}

	recp := e.RcptTo[0]
	addr := strings.ToLower(recp.User + "@" + recp.Host)

	if addr != m.mx.cfg.PublishAddress {
		return backends.NoSuchUser // this server only accepts email for the publish address
	}

	return nil
}

func (m *backend) Initialize(config backends.BackendConfig) error {

	return nil
}

func (m *backend) Reinitialize() error {
	return nil
}

func (m *backend) Shutdown() error {
	return nil
}
