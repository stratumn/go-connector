package main

import (
	"context"
	"fmt"
	"os"
	ossignal "os/signal"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/stratumn/go-node/core"
	event "github.com/stratumn/go-node/core/app/event/service"
	grpcapi "github.com/stratumn/go-node/core/app/grpcapi/service"
	grpcweb "github.com/stratumn/go-node/core/app/grpcweb/service"
	monitoring "github.com/stratumn/go-node/core/app/monitoring/service"
	pruner "github.com/stratumn/go-node/core/app/pruner/service"
	signal "github.com/stratumn/go-node/core/app/signal/service"
	"github.com/stratumn/go-node/core/cfg"
	"github.com/stratumn/go-node/core/manager"

	"github.com/stratumn/go-connector/services/bleveparser"
	"github.com/stratumn/go-connector/services/blevestore"
	"github.com/stratumn/go-connector/services/client"
	"github.com/stratumn/go-connector/services/decryption"
	"github.com/stratumn/go-connector/services/livesync"
	"github.com/stratumn/go-connector/services/logging"
	"github.com/stratumn/go-connector/services/memorystore"
	"github.com/stratumn/go-connector/services/parser"
	"github.com/stratumn/go-connector/services/search"
)

// variables
var (
	Services = []manager.Service{
		&event.Service{},
		&grpcapi.Service{},
		&grpcweb.Service{},
		&monitoring.Service{},
		&pruner.Service{},
		&signal.Service{},
		&client.Service{},
		&decryption.Service{},
		&memorystore.Service{},
		&parser.Service{},
		&livesync.Service{},
		&blevestore.Service{},
		&bleveparser.Service{},
		&search.Service{},
	}

	Config = core.Config{
		BootService:      "boot",
		EnableBootScreen: false,
	}
)

// coreCfgFilename returns the filename of the core config file.
func coreCfgFilename() string {
	return "config.core.toml"
}

// requireCoreConfigSet loads the core's configuration file and exits on failure.
func requireCoreConfigSet() cfg.Set {
	set := core.NewConfigurableSet(Services, &Config)

	// replace the default logging handler
	loggingHandler := &logging.ConfigHandler{}
	set[loggingHandler.ID()] = loggingHandler

	if err := core.LoadConfig(set, coreCfgFilename()); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the core configuration file %q: %s.\n", coreCfgFilename(), err)

		if os.IsNotExist(errors.Cause(err)) {
			fmt.Fprintln(os.Stderr, "You can create one using `stratumn-node init`.")
		}

		os.Exit(1)
	}

	return set
}

func main() {
	config := requireCoreConfigSet().Configs()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 1)

	start := func() {
		c, err := core.New(config, core.OptServices(Services...))
		if err != nil {
			panic(err)
		}

		err = c.Boot(ctx)
		if errors.Cause(err) != context.Canceled {
			panic(err)
		}

		done <- struct{}{}
	}

	go start()

	log.Info("Connector running...")

	// Reload configuration and restart on SIGHUP signal.
	sig := make(chan os.Signal, 1)
	ossignal.Notify(sig, syscall.SIGHUP)

	for range sig {
		cancel()
		<-done
		ctx, cancel = context.WithCancel(context.Background())
		config = requireCoreConfigSet().Configs()
		go start()
	}

	// So the linter doesn't complain.
	cancel()
}
