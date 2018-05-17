package logger

import (
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/sirupsen/logrus"
	"time"
)

// Info logs an info message with context
func Info(logger *logrus.Logger, c *common.Context, message string) {
	logger.WithFields(logrus.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).Info(message)
}

// Debug logs a debug message with context
func Debug(logger *logrus.Logger, c *common.Context, message string) {
	logger.WithFields(logrus.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).Debug(message)
}

// Error logs an error message with context
func Error(logger *logrus.Logger, c *common.Context, message string) {
	logger.WithFields(logrus.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).Error(message)
}
