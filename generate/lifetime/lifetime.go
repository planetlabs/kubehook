package lifetime

import (
	"encoding/json"
	"time"
)

// A Duration is like a time.Duration, except it marshals to a
// time.ParseDuration compatible string.
type Duration int64

// Common durations.
const (
	Nanosecond  Duration = 1
	Microsecond          = 1000 * Nanosecond
	Millisecond          = 1000 * Microsecond
	Second               = 1000 * Millisecond
	Minute               = 60 * Second
	Hour                 = 60 * Minute
)

// String representation of a Duration.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// MarshalJSON marshals a Duration into a time.ParseDuration compatible string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON unmarshals a Duration into from a time.ParseDuration compatible
// string.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(td)
	return nil
}
