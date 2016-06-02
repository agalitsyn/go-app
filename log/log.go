package log

import (
	"os"

	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
)

// Logger provides a leveled-logging interface.
type Logger interface {
	log.Interface
}

// Returns logger instance
func GetLogger(format, level string) Logger {
	if strings.ToLower(format) == "text" {
		log.SetHandler(text.New(os.Stdout))
	} else {
		log.SetHandler(json.New(os.Stdout))
	}
	lvl, err := log.ParseLevel(level)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)

	return getApexLogger()
}

// GetLoggerWithFields returns a logger instance with the specified fields
// without affecting root context.
func GetLoggerWithFields(context string, fields map[string]interface{}) Logger {
	lfields := make(log.Fields, len(fields))
	for key, value := range fields {
		lfields[key] = value
	}
	lfields["context"] = context

	return getApexLogger().WithFields(lfields)
}

// Returns stadrart root logger with additional fields
func getApexLogger() *log.Entry {
	fields := log.Fields{
		"deis.release":  os.Getenv("DEIS_RELEASE"),
		"deis.app":      os.Getenv("DEIS_APP"),
		"deis.hostname": os.Getenv("HOSTNAME"),
	}
	return log.Log.WithFields(fields)
}
