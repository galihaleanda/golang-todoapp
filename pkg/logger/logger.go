package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// New creates a configured logrus logger.
func New(level, env string) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if env == "production" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
	}

	parsed, err := logrus.ParseLevel(level)
	if err != nil {
		parsed = logrus.InfoLevel
	}
	log.SetLevel(parsed)

	return log
}
