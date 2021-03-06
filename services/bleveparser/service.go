package bleveparser

import (
	"context"

	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"

	"github.com/stratumn/go-connector/services/livesync"
)

var (
	// ErrNotStore is returned when the connected service is not a blevestore.
	ErrNotStore = errors.New("connected service is not a blevestore")

	// ErrNotSynchronizer is returned when the connected service is not a synchronizer.
	ErrNotSynchronizer = errors.New("connected service is not a synchronizer")
)

// Service is the Parser service.
type Service struct {
	config *Config

	parser *parser
}

// Config contains configuration options for the Parser service.
type Config struct {
	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the service configuration."`

	// Store is the service used to store the parsed data.
	Store string `toml:"store" comment:"The name of the store service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "bleveparser"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Generic CS bleve Parser"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Generic Chainscript parser listening to new links and storing them in a bleve store"
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Store: "blevestore",
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
	return map[string]struct{}{
		"blevestore": struct{}{},
		"livesync":   struct{}{},
	}
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	s.parser = &parser{}
	if s.parser.idx, ok = exposed[s.config.Store].(bleve.Index); !ok {
		return errors.Wrap(ErrNotStore, s.config.Store)
	}

	if s.parser.synchronizer, ok = exposed["livesync"].(livesync.Synchronizer); !ok {
		return errors.Wrap(ErrNotSynchronizer, "livesync")
	}

	return nil
}

// Expose exposes nothing. No service should ever depend on the parser.
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {

	running()

	errChan := make(chan error)
	go func() {
		err := s.parser.run(ctx)
		errChan <- err
		close(errChan)
	}()

	err := <-errChan
	stopping()

	if err != nil {
		return err
	}
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
			return tree.Set("store", "blevestore")
		},
	}
}
