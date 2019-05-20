package logging

import (
	"io"

	log "github.com/sirupsen/logrus"
)

// WriterHook is a hook that writes logs of specified LogLevels to specified Writer
type WriterHook struct {
	Writer    io.Writer
	LogLevel  log.Level
	Formatter log.Formatter
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (h *WriterHook) Fire(entry *log.Entry) error {
	line, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.Writer.Write(line)
	return err
}

// Levels define on which log levels this hook would trigger
func (h *WriterHook) Levels() []log.Level {
	var levels []log.Level
	for _, l := range log.AllLevels {
		if h.LogLevel >= l {
			levels = append(levels, l)
		}
	}

	return levels
}
