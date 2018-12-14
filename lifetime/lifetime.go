/**
 * Copyright 2018 Planet Labs Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

// ParseDuration parses a string with time.ParseDuration, but returns a
// lifetime.Duration.
func ParseDuration(s string) (Duration, error) {
	td, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return Duration(td), nil
}

// MarshalJSON marshals a Duration into a time.ParseDuration compatible string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON unmarshals a Duration from a time.ParseDuration compatible
// string.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	p, err := ParseDuration(s)
	if err != nil {
		return err
	}

	*d = p
	return nil
}
