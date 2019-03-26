package search

import (
	"context"

	"github.com/blevesearch/bleve"
	logging "github.com/ipfs/go-log"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"
)

// log is the logger for the protocol.
var log = logging.Logger("search")

var (
	// ErrNotStore is returned when the connected service is not a store exposind a DB.
	ErrNotStore = errors.New("connected service is exposing neither a DB not a bleve index")
)

// Service is the Memorystore service.
type Service struct {
	config   *Config
	searcher Searcher
}

// Config contains configuration options for the Memorystore service.
type Config struct {
	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the service configuration."`

	// The name of the store service used for the search.
	Store string `toml:"store" comment:"The name of the store service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "search"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Search"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Basic full-text search"
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
	return map[string]struct{}{s.config.Store: struct{}{}}
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	idx, ok := exposed[s.config.Store].(bleve.Index)
	if !ok {
		return ErrNotStore
	}

	s.searcher = newSearcher(idx)

	return nil
}

// Expose exposes the database client to other services.
// It exposes the database instance.
func (s *Service) Expose() interface{} {
	return s.searcher
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {

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
			return tree.Set("store", "blevestore")
		},
	}
}
