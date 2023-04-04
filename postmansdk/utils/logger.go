package utils

import (
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// GetLogger returns a singleton instance of *logrus.Entry. Use CreateNewLogger
// to create a new logger with a specified log level.
func GetLogger() *logrus.Entry {
	if log != nil {
		return log
	}

	// If by chance someone calls this method without calling CreateNewLogger
	// first. We can always create a new instance of the logger with ErrorLevel
	return CreateNewLogger(logrus.ErrorLevel)
}

// CreateNewLogger returns a new *logrus.Entry object, it overrides
// the previous log level setting, and it is ok to call this function more than once.
func CreateNewLogger(l logrus.Level) *logrus.Entry {
	logger := *logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(l)
	log = logger.WithField("package", "@postman/postman-sdk")

	return log
}
