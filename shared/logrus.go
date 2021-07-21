package shared

import (
	"os"
	"github.com/sirupsen/logrus"
)

type LogrusInfo struct {
	// Version is injected by go (should be a tag name)
	Version string
	// Buildstamp is a timestamp (injected by go) of the build time
	Buildstamp string
	// Githash is the tag for current hash the build represents
	Githash string
}

func SetupLogrus(info LogrusInfo) error {
	var err error
	host, err := os.Hostname()
	if err != nil {
		logrus.Panicln("unable to get Hostname", err)
		return err
	}

	// setup logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.WithFields(logrus.Fields{
		"Version":   info.Version,
		"BuildTime": info.Buildstamp,
		"Githash":   info.Githash,
		"Host":      host,
	}).Info("Service Startup")
	return nil
}
