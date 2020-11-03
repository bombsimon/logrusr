package logrusr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	cases := []struct {
		description  string
		logrusLogger func() logrus.FieldLogger
		logFunc      func(log logr.Logger)
		formatter    func(interface{}) string
		assertions   map[string]string
	}{
		{
			description: "basic logging",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			assertions: map[string]string{
				"level": "info",
				"msg":   "hello, world",
			},
		},
		{
			description: "set name once",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.WithName("main").Info("hello, world")
			},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"logger": "main",
			},
		},
		{
			description: "set name twice",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.WithName("main").WithName("subpackage").Info("hello, world")
			},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"logger": "main.subpackage",
			},
		},
		{
			description: "V(0) logging with info level set is shown",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.V(0).Info("hello, world")
			},
			assertions: map[string]string{
				"level": "info",
				"msg":   "hello, world",
			},
		},
		{
			description: "V(2) logging with info level set is not shown",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.V(2).Info("hello, world")
			},
			assertions: nil,
		},
		{
			description: "V(2) logging with trace level set is shown",
			logrusLogger: func() logrus.FieldLogger {
				l := logrus.New()
				l.SetLevel(logrus.TraceLevel)

				return l
			},
			logFunc: func(log logr.Logger) {
				log.V(2).Info("hello, world")
			},
			assertions: map[string]string{
				"level": "info",
				"msg":   "hello, world",
			},
		},
		{
			description: "arguments are added while calling Info()",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", "walrus")
			},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"animal": "walrus",
			},
		},
		{
			description: "arguments are added after WithValues()",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.WithValues("color", "green").Info("hello, world", "animal", "walrus")
			},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"animal": "walrus",
				"color":  "green",
			},
		},
		{
			description: "error logs have the approperate information",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.Error(errors.New("this is error"), "error occurred")
			},
			assertions: map[string]string{
				"level": "error",
				"msg":   "error occurred",
				"error": "this is error",
			},
		},
		{
			description: "error shown with lov severity logger",
			logrusLogger: func() logrus.FieldLogger {
				l := logrus.New()
				l.SetLevel(logrus.ErrorLevel)

				return l
			},
			logFunc: func(log logr.Logger) {
				log.Error(errors.New("this is error"), "error occurred")
			},
			assertions: map[string]string{
				"level": "error",
				"msg":   "error occurred",
				"error": "this is error",
			},
		},
		{
			description: "bad number of arguments discards all",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", "walrus", "foo")
			},
			assertions: map[string]string{
				"level":   "info",
				"msg":     "hello, world",
				"-animal": "walrus",
			},
		},
		{
			description: "complex data types are converted",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "animal", []byte("walrus"), "list", []int{1, 2, 3})
			},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"animal": "walrus",
				"list":   "[1,2,3]",
			},
		},
		{
			description: "custom formatter is used",
			logrusLogger: func() logrus.FieldLogger {
				return logrus.New()
			},
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "list", []int{1, 2, 3})
			},
			formatter: func(val interface{}) string {
				return fmt.Sprintf("%v", val)
			},
			assertions: map[string]string{
				"level": "info",
				"msg":   "hello, world",
				"list":  "[1 2 3]",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			// Use a buffer for our output.
			logWriter := &bytes.Buffer{}

			// Create the logger according to the test case.
			logrusLogger := tc.logrusLogger()

			switch l := logrusLogger.(type) {
			case *logrus.Logger:
				l.SetOutput(logWriter)
				l.SetFormatter(&logrus.JSONFormatter{})
			case *logrus.Entry:
				l.Logger.SetOutput(logWriter)
				l.Logger.SetFormatter(&logrus.JSONFormatter{})
			default:
				t.FailNow()
			}

			// Send the created logger to the test case to invoke desired
			// logging.
			tc.logFunc(NewLoggerWithFormatter(logrusLogger, tc.formatter))

			if tc.assertions == nil {
				assert.Equal(t, logWriter.Len(), 0)
				return
			}

			var loggedLine map[string]string
			err := json.Unmarshal(logWriter.Bytes(), &loggedLine)

			require.NoError(t, err)

			for k, v := range tc.assertions {
				field, ok := loggedLine[k]

				// Annotate negative tests with a minus. To ensure `key` is
				// *not* in the output, name the assertion `-key`.
				if strings.HasPrefix(k, "-") {
					assert.False(t, ok)
					assert.Empty(t, field)

					continue
				}

				assert.True(t, ok)
				assert.NotEmpty(t, field)
				assert.Equal(t, v, field)
			}
		})
	}
}
