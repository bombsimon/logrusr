package logrusr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description  string
		logrusLevel  logrus.Level
		logFunc      func(log logr.Logger)
		formatter    FormatFunc
		reportCaller bool
		defaultName  []string
		assertions   map[string]string
	}{
		{
			description: "basic logging",
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
			description: "set name and values and name again",
			logFunc: func(log logr.Logger) {
				log.
					WithName("main").
					WithValues("k1", "v1", "k2", "v2").
					WithName("subpackage").
					Info("hello, world", "k3", "v3")
			},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"logger": "main.subpackage",
				"k1":     "v1",
				"k2":     "v2",
				"k3":     "v3",
			},
		},
		{
			description: "V(0) logging with info level set is shown",
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
			logFunc: func(log logr.Logger) {
				log.V(1).Info("hello, world")
				log.V(2).Info("hello, world")
			},
			assertions: nil,
		},
		{
			description: "V(1) logging with debug level set is shown",
			logrusLevel: logrus.DebugLevel,
			logFunc: func(log logr.Logger) {
				log.V(1).Info("hello, world")
			},
			assertions: map[string]string{
				"level": "debug",
				"msg":   "hello, world",
			},
		},
		{
			description: "V(2) logging with trace level set is shown",
			logrusLevel: logrus.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(2).Info("hello, world")
			},
			assertions: map[string]string{
				"level": "trace",
				"msg":   "hello, world",
			},
		},
		{
			description: "negative V-logging truncates to info",
			logrusLevel: logrus.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(-10).Info("hello, world")
			},
			assertions: map[string]string{
				"level": "info",
				"msg":   "hello, world",
			},
		},
		{
			description: "additive V-logging, negatives ignored",
			logrusLevel: logrus.TraceLevel,
			logFunc: func(log logr.Logger) {
				log.V(0).V(1).V(-20).V(1).Info("hello, world")
			},
			assertions: map[string]string{
				"level": "trace",
				"msg":   "hello, world",
			},
		},
		{
			description: "arguments are added while calling Info()",
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
			description: "error logs have the appropriate information",
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
			logFunc: func(log logr.Logger) {
				log.Info("hello, world", "list", []int{1, 2, 3})
			},
			formatter: func(val interface{}) interface{} {
				return fmt.Sprintf("%v", val)
			},
			assertions: map[string]string{
				"level": "info",
				"msg":   "hello, world",
				"list":  "[1 2 3]",
			},
		},
		{
			description: "with default name",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			defaultName: []string{"some", "name"},
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"logger": "some.name",
			},
		},
		{
			description: "without report caller",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			assertions: map[string]string{
				"level":   "info",
				"msg":     "hello, world",
				"-caller": "no-caller",
			},
		},
		{
			description: "with report caller",
			logFunc: func(log logr.Logger) {
				log.Info("hello, world")
			},
			reportCaller: true,
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"caller": `~logrusr_test.go:\d+`,
			},
		},
		{
			description: "with report caller and depth",
			logFunc: func(log logr.Logger) {
				log.WithCallDepth(2).Info("hello, world")
			},
			reportCaller: true,
			assertions: map[string]string{
				"level":  "info",
				"msg":    "hello, world",
				"caller": `~testing.go:\d+`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc := tc
			t.Parallel()

			// Use a buffer for our output.
			logWriter := &bytes.Buffer{}

			logrusLogger := logrus.New()
			logrusLogger.SetOutput(logWriter)
			logrusLogger.SetFormatter(&logrus.JSONFormatter{})

			if tc.logrusLevel != logrus.PanicLevel {
				logrusLogger.SetLevel(tc.logrusLevel)
			}

			// Send the created logger to the test case to invoke desired
			// logging.
			opts := []Option{
				WithFormatter(tc.formatter),
			}

			if tc.reportCaller {
				opts = append(opts, WithReportCaller())
			}

			if tc.defaultName != nil {
				opts = append(opts, WithName(tc.defaultName...))
			}

			tc.logFunc(New(logrusLogger, opts...))

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

				// Annotate regexp matches with the value starting with a tilde
				// (~). The tilde will be dropped and used to compile a regexp to
				// match the field.
				if strings.HasPrefix(v, "~") {
					assert.Regexp(t, regexp.MustCompile(v[1:]), field)
					continue
				}

				assert.True(t, ok)
				assert.NotEmpty(t, field)
				assert.Equal(t, v, field)
			}
		})
	}
}
