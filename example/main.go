package main

import (
	"errors"
	"fmt"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		log logr.Logger
	)

	logrusLog := logrus.New()

	log = logrusr.NewLogger(logrusLog)
	log = log.WithName("MyName").WithValues("user", "you")
	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(0).Info("you should see this")
	log.V(2).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(errors.New("caught error"), "goodbye", "code", -1)

	fmt.Println("")

	entryLog := logrus.New().WithFields(logrus.Fields{
		"some_field":    "some_value",
		"another_field": 42,
	})

	log = logrusr.NewLogger(entryLog)
	log = log.WithName("MyName").WithValues("user", "you")
	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(0).Info("you should see this")
	log.V(2).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(errors.New("caught error"), "goodbye", "code", -1)

	log = log.WithName("subpackage")
	log.Info("hello from subpackage")
}
