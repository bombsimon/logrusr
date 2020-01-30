# Logrusr

A [logr](https://github.com/go-logr/logr) implementation using
[logrus](https://github.com/sirupsen/logrus).

## Usage

```go
import (
    "github.com/sirupsen/logrus"
    "github.com/bombsimon/logrusr"
    "github.com/go-logr/logr"
)

func main() {
    var log logr.Logger

    logger = logrus.New()
    log = logrusr.NewLogger(logger)

    log.Info("Logr in action!", "the answer", 42)
}
```

For more details, see [example](example/main.go).

## Implementation details

The NewLogger method takes a `logrus.FieldLogger` interface as input which means
this works with both `logrus.Logger` and `logrus.Entry`. This is currently a
quite naive implementation in early state. Use with caution.
