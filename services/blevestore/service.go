package blevestore

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("service", "blevestore")

// Service is the Blevestore service.
type Service struct {
	config *Config

	store *store
}

// Config contains configuration options for the Blevestore service.
type Config struct {
	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the service configuration."`

	// Path is the path to the bleve store data.
	Path string `toml:"path" comment:"The path to the bleve store data."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "blevestore"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Bleve Store"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "In-Memory Search index"
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

// Expose exposes the database client to other services.
// It exposes the database instance.
func (s *Service) Expose() interface{} {
	return s.store.idx
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	var err error
	s.store, err = newStore(s.config.Path)
	if err != nil {
		return err
	}

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
			return tree.Set("path", "bleve_store")
		},
	}
}
