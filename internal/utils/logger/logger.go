package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// New - create new logger with log level and output log file directory
func New(logLevel uint, logDir string, format logrus.Formatter) (*logrus.Logger, error) {
	log := logrus.New()
	log.Level = logrus.Level(logLevel)

	var logFName string
	ts := time.Now().Format("tccbot-2006-01-02-15-04-05")
	if logDir != "" {
		logFName = filepath.Join(logDir, ts+".log")

		logFile, err := os.OpenFile(logFName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		log.SetOutput(logFile)
		logrus.RegisterExitHandler(func() {
			if logFile == nil {
				return
			}
			logFile.Close()
		})
	}

	log.SetFormatter(format)

	log.WithFields(logrus.Fields{"application": "tccbot", "location": logFName})

	return log, nil
}
