package logger

import (
	"github.com/lawrencegripper/ion/internal/pkg/common"
	log "github.com/sirupsen/logrus"
	"time"
)

// Info logs an info message with context
func Info(c *common.Context, message string) {
	log.WithFields(log.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).Info(message)
}

// Debug logs a debug message with context
func Debug(c *common.Context, message string) {
	log.WithFields(log.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).Debug(message)
}

// Error logs an error message with context
func Error(c *common.Context, message string) {
	log.WithFields(log.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).Error(message)
}
