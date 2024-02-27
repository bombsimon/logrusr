//go:build go1.21
// +build go1.21

package logrusr_test

import (
	"errors"
	"log/slog"
	"os"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func ExampleNew() {
	l := logrus.New()
	l.Formatter.(*logrus.TextFormatter).DisableTimestamp = true
	l.Out = os.Stdout
	log := slog.New(logr.ToSlogHandler(logrusr.New(l)))
	log = log.With("number", slog.IntValue(123))
	log.Debug("do not print this!")
	log.Info("print this please", slog.String("key", "value"), slog.Group("test", slog.Group("nested", "foo", "bar")))
	log.Error("oh no", "error", errors.New("some error"))
	// Output:
	// level=info msg="print this please" key=value number=123 test.nested.foo=bar
	// level=error msg="oh no" error="some error" number=123
}
