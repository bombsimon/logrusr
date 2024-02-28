//go:build !go1.21
// +build !go1.21

package logrusr

import (
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

		if v, ok := v.(logr.Marshaler); ok {
			f[s] = v.MarshalLog()
			continue
		}

		// Try to avoid marshaling known types.
		switch vVal := v.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64, complex64, complex128,
			string, bool, error:
			f[s] = vVal

		case []byte:
			f[s] = string(vVal)

		default:
			f[s] = formatterOrMarshal(v, formatter)
		}
	}

	return f
}
