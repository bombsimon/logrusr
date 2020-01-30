package logrusr

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// According to the specification of the Logger interface calling the InfoLogger
// directly on the logger should be the same as calling them on V(0). Since
// logrus level 0 is PanicLevel and Infolevel doesn't start until V(4) we use
// this constant to be able to calculate what V(n) values should mean.
const (
	logrusDiffToInfo = 4
)

type logrusr struct {
	logger logrus.FieldLogger
}

// NewLogger will return a new logr.Logger from a logrus.FieldLogger.
func NewLogger(l logrus.FieldLogger) logr.Logger {
	return &logrusr{
		logger: l,
	}
}

// Enabled is a part of the InfoLogger interface. It will return true if the
// logrus.Logger has a level set to logrus.InfoLevel or higher (Debug/Trace).
func (l *logrusr) Enabled() bool {
	var log *logrus.Logger

	switch t := l.logger.(type) {
	case *logrus.Logger:
		log = t

	case *logrus.Entry:
		log = t.Logger
	}

	return log.GetLevel() <= logrus.InfoLevel
}

// V is a part of the Logger interface. It will create a new instance of a
// *logrus.Logger instead of changing the current one and the new logger will be
// retruned. According to the documentation level V(0) should be equivalent as
// calling Info() directly on the logger. To ensure this the constant
// `logrusDiffToInfo` will be added to all passed values so that V(0) creates a
// logger with level logrus.InfoLevel and V(2) would create a logger with level
// logrus.TraceLevel.
func (l *logrusr) V(level int) logr.InfoLogger {
	var (
		newLogger = logrus.New()
		oldLogger *logrus.Logger
		oldFields logrus.Fields
	)

	switch log := l.logger.(type) {
	case *logrus.Entry:
		oldLogger = log.Logger
		oldFields = log.Data

	case *logrus.Logger:
		oldLogger = log
	}

	// To make V(0) set the verbosity on the logger to InfoLevel instead of
	// PanicLevel which is 0-level.
	logrusLevel := logrus.Level(level + logrusDiffToInfo)
	newLogger.SetLevel(logrusLevel)

	// Keep the formatter used on the old logger.
	newLogger.SetFormatter(oldLogger.Formatter)

	// Add fields from the old logger (if there were any)
	newLoggerWithFields := newLogger.WithFields(oldFields)

	return NewLogger(newLoggerWithFields)
}

// WithValues is a part of the Logger interface. This is equivalent to
// logrus WithFields() but takes a list of even arguments (key/value pairs)
// instead of a map as input. If an odd number of arguments are sent all values
// will be discarded.
func (l *logrusr) WithValues(keysAndValues ...interface{}) logr.Logger {
	newLogger := l.logger
	newFields := listToLogrusFields(keysAndValues...)

	return NewLogger(newLogger.WithFields(newFields))
}

// WithName is a part of the Logger interface. This will set the key "name" as a
// logrus field.
func (l *logrusr) WithName(name string) logr.Logger {
	l.logger = l.logger.WithFields(logrus.Fields{
		"name": name,
	})

	return l
}

// Info logs info messages if the logger is enabled, that is if the level on the
// logger is set to logrus.InfoLevel or less.
func (l *logrusr) Info(msg string, keysAndValues ...interface{}) {
	if !l.Enabled() {
		return
	}

	l.logger.
		WithFields(listToLogrusFields(keysAndValues...)).
		Info(msg)
}

// Error logs error messages.
func (l *logrusr) Error(err error, msg string, keysAndValues ...interface{}) {
	l.logger.
		WithFields(listToLogrusFields(keysAndValues...)).
		WithError(err).
		Error(msg)
}

// listToLogrusFields converts a list of arbitrary length to key/value paris.
func listToLogrusFields(keysAndValues ...interface{}) logrus.Fields {
	var f = logrus.Fields{}

	// Skip all fields if it's not an even lengthed list.
	if len(keysAndValues)%2 != 0 {
		return f
	}

	for i := 0; i < len(keysAndValues); i += 2 {
		k, v := keysAndValues[i], keysAndValues[i+1]

		if s, ok := k.(string); ok {
			// Try to avoid marshaling known types.
			switch vVal := v.(type) {
			case int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64,
				float32, float64, complex64, complex128,
				string, bool:
				f[s] = vVal

			case []byte:
				f[s] = string(vVal)

			default:
				j, _ := json.Marshal(vVal)
				f[s] = string(j)
			}
		}
	}

	return f
}
