package main

import (
	"errors"
	"fmt"

	"github.com/bombsimon/logrusr/v4"
	"github.com/sirupsen/logrus"
)

type customType string

func (c customType) MarshalLog() any {
	return "custom marshal message"
}

func main() {
	logrusLog := logrus.New()
	log := logrusr.New(logrusLog)

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
	log = logrusr.New(entryLog)

	log = log.WithName("MyName").WithValues("user", "you")
	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(0).Info("you should see this")
	log.V(2).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(errors.New("caught error"), "goodbye", "code", -1)

	fmt.Println("")

	log = log.WithName("subpackage")
	log.Info("hello from subpackage")
	log.WithName("even_deeper").Info("hello even deeper")

	fmt.Println("")

	logrusLog = logrus.New()
	logrusLog.SetLevel(logrus.TraceLevel)

	log = logrusr.New(
		logrusLog,
		logrusr.WithReportCaller(),
	)

	log.V(0).Info("you should see this as info")
	log.V(1).Info("you should see this as debug")
	log.V(2).Info("you should see this as trace")
	log.V(1).V(1).Info("you should see this as trace")
	log.V(10).Info("you should not see this")

	fmt.Println("")

	type myInt int

	logrusLog = logrus.New()
	logrusLog.SetFormatter(&logrus.JSONFormatter{})

	log = logrusr.New(
		logrusLog,
		logrusr.WithFormatter(func(in any) any {
			if v, ok := in.(myInt); ok && v == 1 {
				return "1"
			}

			return in
		}),
	)

	log.Info("custom types", "my_int", myInt(1), "my_other_int", myInt(2))

	myString := customType("original value")
	log.Info("custom marshaller", "my_string", myString)
}
