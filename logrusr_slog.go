//go:build go1.21
// +build go1.21

package logrusr

import (
	"encoding/json"
	"log/slog"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// listToLogrusFields converts a list of arbitrary length to key/value paris.
func listToLogrusFields(formatter FormatFunc, keysAndValues ...interface{}) logrus.Fields {
	f := make(logrus.Fields)

	// Skip all fields if it's not an even length list.
	if len(keysAndValues)%2 != 0 {
		return f
	}

	for i := 0; i < len(keysAndValues); i += 2 {
		k, v := keysAndValues[i], keysAndValues[i+1]

		s, ok := k.(string)
		if !ok {
			continue
		}

		// Try to avoid marshaling known types.
		switch vVal := v.(type) {
		case logr.Marshaler:
			f[s] = vVal.MarshalLog()

		case slog.LogValuer:
			f[s] = slog.AnyValue(vVal).Resolve()

		case slog.Value:
			formatSlog(f, formatter, slog.Attr{Key: s, Value: vVal})

		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64, complex64, complex128,
			string, bool, error:
			f[s] = vVal

		case []byte:
			f[s] = string(vVal)

		default:
			if formatter != nil {
				f[s] = formatter(v)
			} else {
				j, _ := json.Marshal(v)
				f[s] = string(j)
			}
		}
	}

	return f
}

func formatSlog(fields logrus.Fields, formatter FormatFunc, attr slog.Attr) {
	switch attr.Value.Kind() {
	case slog.KindString:
		fields[attr.Key] = attr.Value.String()

	case slog.KindLogValuer:
		formatSlog(fields, formatter, slog.Attr{
			Key:   attr.Key,
			Value: attr.Value.Resolve(),
		})

	case slog.KindInt64:
		fields[attr.Key] = attr.Value.Int64()

	case slog.KindUint64:
		fields[attr.Key] = attr.Value.Uint64()

	case slog.KindFloat64:
		fields[attr.Key] = attr.Value.Float64()

	case slog.KindBool:
		fields[attr.Key] = attr.Value.Bool()

	case slog.KindDuration:
		fields[attr.Key] = attr.Value.Duration()

	case slog.KindTime:
		fields[attr.Key] = attr.Value.Time()

	case slog.KindGroup:
		attrs := attr.Value.Group()
		if attr.Key == "" {
			// Inline group
			for _, attr := range attrs {
				formatSlog(fields, formatter, attr)
			}
			return
		}
		if len(attrs) == 0 {
			return
		}
		value := make(logrus.Fields)
		for _, attr := range attrs {
			formatSlog(value, formatter, attr)
		}
		fields[attr.Key] = value

	default:
		if formatter != nil {
			fields[attr.Key] = formatter(attr.Value.Any())
		}
		v, _ := json.Marshal(attr.Value.Any())
		fields[attr.Key] = string(v)
	}
}
