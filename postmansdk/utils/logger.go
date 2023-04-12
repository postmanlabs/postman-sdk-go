package utils

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Entry

// CreateNewLogger returns a new *logrus.Entry object, it overrides
// the previous log level setting, and it is ok to call this function more than once.
func CreateNewLogger(l logrus.Level) {
	logger := *logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(l)
	Log = logger.WithField("package", "@postman/postman-sdk")
}
