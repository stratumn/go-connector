package decryption

import (
	"context"

	logging "github.com/ipfs/go-log"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"
)

var log = logging.Logger("decryption")

// Service is the Ping service.
type Service struct {
	config *Config

	decryptor Decryptor
}

// Config contains configuration options for the Ping service.
type Config struct {
	// SigningPrivateKey is pretty well named.
	EncryptionPrivateKey string `toml:"encryption_private_key" comment:"The encryption private key."`

	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the service configuration."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "decryption"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Decryption"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Service used to decrypt links."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	return nil
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	return nil
}

// Expose exposes the stratumn client to other services.
//
// It exposes the decryptor instance
func (s *Service) Expose() interface{} {
	return s.decryptor
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	d, err := newDecryptor([]byte(s.config.EncryptionPrivateKey))
	if err != nil {
		return err
	}
	s.decryptor = d

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
			return tree.Set("encryption_private_key", "")
		},
	}
}
