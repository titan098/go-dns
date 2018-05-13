package logging

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// SetupLogging create the logging context and exposes the log
func SetupLogging(module string) *logrus.Logger {
	log := logrus.New()

	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true

	log.Formatter = formatter
	log.Level = logrus.DebugLevel
	return log
}
