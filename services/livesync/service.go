package livesync

import (
	"context"
	"time"

	"github.com/stratumn/go-connector/services/client"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/cfg"
)

// DefaultPollInterval is the default interval at which the livesync service calls Startumn APIs
const DefaultPollInterval = 10 * time.Second

var (
	// ErrNotClient is returned when the connected service is not a stratumn client.
	ErrNotClient = errors.New("connected service is not a stratumn client")
)

// Service is the Livesync service.
type Service struct {
	config *Config

	synchronizer *synchronizer
}

// Config contains configuration options for the Livesync service.
type Config struct {
	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the service configuration."`

	PollInterval     time.Duration `toml:"poll_interval" comment:"The frenquency at which the livesync service polls data from Stratumn APIs."`
	WatchedWorkflows []uint        `toml:"watched_workflows" comment:"The IDs of the workflows to synchronize data from."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "livesync"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Stratumn API Live Sync"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Polls data from the Stratumn APIs at regular intervals"
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		PollInterval: DefaultPollInterval,
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
		"stratumnClient": struct{}{},
	}
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool
	var stratumnClient client.StratumnClient
	if stratumnClient, ok = exposed["stratumnClient"].(client.StratumnClient); !ok {
		return errors.Wrap(ErrNotClient, "stratumnClient")
	}

	s.synchronizer = newSycnhronizer(stratumnClient, s.config.WatchedWorkflows)

	return nil
}

// Expose exposes the synchronizer client to other services.
// It exposes the Synchronizer instance.
func (s *Service) Expose() interface{} {
	return s.synchronizer
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	ticker := time.NewTicker(time.Second * s.config.PollInterval)

	running()

RUN_LOOP:
	for {
		select {
		case <-ticker.C:
			err := s.synchronizer.pollAndNotify(ctx)
			if err != nil {
				stopping()
				return err
			}
		case <-ctx.Done():
			break RUN_LOOP
		}
	}

	ticker.Stop()
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
			return tree.Set("poll_interval", 1)
		},
	}
}
