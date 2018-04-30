package logging

import (
	"github.com/sirupsen/logrus"
)

// SetupLogging create the logging context and exposes the log
func SetupLogging(module string) *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return log
}
