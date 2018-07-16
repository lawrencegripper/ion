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

// InfoWithFields logs an info message with context
func InfoWithFields(c *common.Context, message string, fields map[string]interface{}) {
	log.WithFields(log.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).WithFields(fields).Info(message)
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

// DebugWithFields logs an info message with context
func DebugWithFields(c *common.Context, message string, fields map[string]interface{}) {
	log.WithFields(log.Fields{
		"eventID":       c.EventID,
		"correlationID": c.CorrelationID,
		"name":          c.Name,
		"timestamp":     time.Now(),
	}).WithFields(fields).Debug(message)
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
