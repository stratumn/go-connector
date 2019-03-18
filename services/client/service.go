package client

import (
	"context"

	"go-connector/services/decryption"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"
)

var (
	// ErrNotDecryptor is returned when the connected service is not a decryptor.
	ErrNotDecryptor = errors.New("connected service is not a decryptor")
)

// Service is the Ping service.
type Service struct {
	config *Config
	client StratumnClient

	decryptor decryption.Decryptor
}

// Config contains configuration options for the Ping service.
type Config struct {
	// TraceUrl is the URL to trace.
	TraceURL string `toml:"trace_url" comment:"The URL of Stratumn Trace APIs."`
	// AccountUrl is the URL to account.
	AccountURL string `toml:"account_url" comment:"The URL of Stratumn Account APIs."`

	// SigningPrivateKey is pretty well named.
	SigningPrivateKey string `toml:"signing_private_key" comment:"The signing private key."`

	// The name of the decryption service.
	Decryption string `toml:"decryption" comment:"The name of the decryption service."`

	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the service configuration."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "stratumnClient"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Stratumn Client"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Client to Stratumn services APIs."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		TraceURL:   "https://trace-api.stratumn.com",
		AccountURL: "https://account-api.stratumn.com",
		Decryption: "decryption",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Decryption] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.decryptor, ok = exposed[s.config.Decryption].(decryption.Decryptor); !ok {
		return errors.Wrap(ErrNotDecryptor, s.config.Decryption)
	}

	return nil
}

// Expose exposes the stratumn client to other services.
//
// It exposes the Stratumn client.
func (s *Service) Expose() interface{} {
	return s.client
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.client = newClient(s.config.TraceURL, s.config.AccountURL, []byte(s.config.SigningPrivateKey), s.decryptor)

	running()
	<-ctx.Done()
	stopping()

	return errors.WithStack(ctx.Err())
}

// Migrator methods.

// VersionKey is the version key.
func (s *Service) VersionKey() string {
	return "configuration_version"
}

// Migrations is the services migrations.
func (s *Service) Migrations() []cfg.MigrateHandler {
	return []cfg.MigrateHandler{
		func(tree *cfg.Tree) error {
			err := tree.Set("trace_url", "https://trace-api.staging.stratumn.rocks")
			if err != nil {
				return err
			}
			err = tree.Set("account_url", "https://account-api.staging.stratumn.rocks")
			if err != nil {
				return err
			}
			err = tree.Set("signing_private_key", "")
			if err != nil {
				return err
			}
			return tree.Set("decryption", "decryption")
		},
	}
}
