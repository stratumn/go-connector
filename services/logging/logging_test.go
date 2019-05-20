package logging_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-connector/services/logging"
)

func TestConfigHandler(t *testing.T) {

	t.Run("configures logrus", func(t *testing.T) {
		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		writers := []logging.WriterConfig{
			logging.WriterConfig{
				Type:      logging.File,
				Filename:  f.Name(),
				Formatter: logging.JSON,
				Level:     "debug",
			},
			logging.WriterConfig{
				Type:      logging.Stdout,
				Formatter: logging.Text,
				Level:     "info",
			},
			logging.WriterConfig{
				Type:      logging.Stderr,
				Formatter: logging.Text,
				Level:     "error",
			},
		}
		c := logging.Config{Writers: writers}
		handler := logging.ConfigHandler{}

		err = handler.SetConfig(c)
		require.NoError(t, err)

		logger := logrus.StandardLogger()
		assert.Equal(t, ioutil.Discard, logger.Out)
		assert.Equal(t, logrus.TraceLevel, logger.Level)

		assert.Len(t, logger.Hooks[logrus.DebugLevel], 1)
		assert.Len(t, logger.Hooks[logrus.InfoLevel], 2)
		assert.Len(t, logger.Hooks[logrus.ErrorLevel], 3)

		assert.EqualValues(t, logger.Hooks[logrus.InfoLevel][1], &logging.WriterHook{Formatter: &logrus.TextFormatter{ForceColors: true, DisableColors: false}, Writer: os.Stdout, LogLevel: logrus.InfoLevel})
		assert.EqualValues(t, logger.Hooks[logrus.ErrorLevel][2], &logging.WriterHook{Formatter: &logrus.TextFormatter{ForceColors: true, DisableColors: false}, Writer: os.Stderr, LogLevel: logrus.ErrorLevel})
	})

	t.Run("setting the config twice does not duplicate hooks", func(t *testing.T) {
		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		writers := []logging.WriterConfig{
			logging.WriterConfig{
				Type:      logging.File,
				Filename:  f.Name(),
				Formatter: logging.JSON,
				Level:     "debug",
			},
			logging.WriterConfig{
				Type:      logging.Stdout,
				Formatter: logging.Text,
				Level:     "info",
			},
			logging.WriterConfig{
				Type:      logging.Stderr,
				Formatter: logging.Text,
				Level:     "error",
			},
		}
		c := logging.Config{Writers: writers}
		handler := logging.ConfigHandler{}

		err = handler.SetConfig(c)
		require.NoError(t, err)

		// re-set the config
		err = handler.SetConfig(c)
		require.NoError(t, err)

		logger := logrus.StandardLogger()
		assert.Equal(t, ioutil.Discard, logger.Out)
		assert.Equal(t, logrus.TraceLevel, logger.Level)

		assert.Len(t, logger.Hooks[logrus.DebugLevel], 1)
		assert.Len(t, logger.Hooks[logrus.InfoLevel], 2)
		assert.Len(t, logger.Hooks[logrus.ErrorLevel], 3)
	})
}
