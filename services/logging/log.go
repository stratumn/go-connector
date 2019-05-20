package logging

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// ErrInvalidWriterType is returned when a writer has an invalid type.
	ErrInvalidWriterType = errors.New("log writer has an invalid type")

	// ErrInvalidWriterLevel is returned when a writer has an invalid
	// level.
	ErrInvalidWriterLevel = errors.New("log writer has an invalid level")

	// ErrInvalidWriterFormatter is returned when a writer has an invalid
	// formatter.
	ErrInvalidWriterFormatter = errors.New("log writer has an invalid formatter")
)

// Log writers.
const (
	File   = "file"
	Stdout = "stdout"
	Stderr = "stderr"
)

// Log formatters.
const (
	JSON = "json"
	Text = "text"
)

// WriterConfig contains configuration for a log writer.
type WriterConfig struct {
	// Type is the type of the writer.
	Type string `toml:"type" comment:"The type of writer (file, stdout, stderr)."`

	// Level is the log level for the writer.
	Level string `toml:"level" comment:"The log level for the writer (trace, debug, info, warn, error, fatal, panic)."`

	// File is the log formatter for the writer.
	Formatter string `toml:"formatter" comment:"The formatter for the writer (json, text)."`

	// Filename is the file for a file logger.
	Filename string `toml:"filename" comment:"The file for a file logger."`
}

// Config contains configuration options for the log.
type Config struct {
	// Writers are the writers for the log.
	Writers []WriterConfig `toml:"writers"`
}

// ConfigHandler is the handler of the log configuration.
type ConfigHandler struct {
	config *Config
}

// ID returns the unique identifier of the configuration.
func (h *ConfigHandler) ID() string {
	return "log"
}

// Config returns the current service configuration or creates one with
// good default values.
//
// The default configuration writes to a single file using the JSON formatter
// and weekly rotations.
func (h *ConfigHandler) Config() interface{} {
	if h.config != nil {
		return *h.config
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.WithStack(err))
	}

	filename, err := filepath.Abs(filepath.Join(cwd, "log.jsonld"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	return Config{
		Writers: []WriterConfig{{
			Type:      File,
			Level:     "debug",
			Formatter: JSON,
			Filename:  filename,
		}},
	}
}

// SetConfig configures the service handler.
func (h *ConfigHandler) SetConfig(config interface{}) error {
	conf := config.(Config)
	h.config = &conf

	// we need to reset all the hooks to avoid duplicate when setting the config twice.
	logger := logrus.StandardLogger()
	for level := range logger.Hooks {
		logger.Hooks[level] = []logrus.Hook{}
	}

	logrus.SetOutput(ioutil.Discard)   // Send all logs to nowhere by default
	logrus.SetLevel(logrus.TraceLevel) // Set the logging level to the minimum

	for _, writerConf := range conf.Writers {
		logLevel, err := logrus.ParseLevel(writerConf.Level)
		if err != nil {
			return err
		}

		var w io.Writer
		switch writerConf.Type {
		case Stdout:
			w = os.Stdout
		case Stderr:
			w = os.Stderr
		case File:
			w, err = os.OpenFile(writerConf.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				return err
			}
		default:
			return errors.WithStack(ErrInvalidWriterType)
		}

		var formatter logrus.Formatter
		switch writerConf.Formatter {
		case Text:
			formatter = &logrus.TextFormatter{
				ForceColors:      true,
				DisableTimestamp: false,
			}
		case JSON:
			formatter = &logrus.JSONFormatter{}
		default:
			return errors.WithStack(ErrInvalidWriterFormatter)
		}

		// Send logs with level higher than `logLevel` to the appropriate writer and formatter.
		logrus.AddHook(&WriterHook{
			Writer:    w,
			LogLevel:  logLevel,
			Formatter: formatter,
		})

	}

	return nil
}
