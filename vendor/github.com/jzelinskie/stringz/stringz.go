// Copyright 2019 Jimmy Zelinskie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package stringz implements a collection of utility functions for
// manipulating strings and lists of strings.
package stringz

import "strings"

// SliceContains returns true if the provided string is in the provided string
// slice.
func SliceContains(ys []string, x string) bool {
	for _, y := range ys {
		if x == y {
			return true
		}
	}
	return false
}

// Dedup returns a new slice with any duplicates removed.
func Dedup(xs []string) []string {
	set := make(map[string]struct{}, 0)
	ys := make([]string, 0, len(xs))
	for _, x := range xs {
		if _, alreadyExists := set[x]; alreadyExists {
			continue
		}
		ys = append(ys, x)
		set[x] = struct{}{}
	}

	return ys
}

// DefaultEmpty returns the fallback when val is empty string.
//
// This function is inspired by Python's `dict.get()`.
func DefaultEmpty(val, fallback string) string {
	return Default(val, fallback, "")
}

// Default returns a fallback value when the provided value is equal to any
// of the zero values.
func Default(val, fallback string, zeroValues ...string) string {
	for _, zeroValue := range zeroValues {
		if val == zeroValue {
			return fallback
		}
	}

	return val
}

// SliceEqual returns true if two string slices are the same.
func SliceEqual(xs, ys []string) bool {
	if len(xs) != len(ys) {
		return false
	}

	for i, x := range xs {
		if x != ys[i] {
			return false
		}
	}

	return true
}

// TrimPrefixIndex trims everything before the provided index.
func TrimPrefixIndex(s, index string) string {
	i := strings.Index(s, index)
	if i <= 0 {
		return s
	}
	return s[i+len(index):]
}

// TrimSurrounding returns a string with both a prefix and suffix trimmed from
// it.
//
// Do not confuse this with strings.Trim() which removes characters in a cutset
// rather than working on prefixes and suffixes.
func TrimSurrounding(s, surrounding string) string {
	s = strings.TrimPrefix(s, surrounding)
	return strings.TrimSuffix(s, surrounding)
}

// SliceMap is a functional-style mapping function for slices of strings.
//
// This is particularly useful when you would normally use a for-loop, but want
// `defer` to execute for each iteration.
func SliceMap(xs []string, fn func(string) error) error {
	for _, x := range xs {
		err := fn(x)
		if err != nil {
			return err
		}
	}
	return nil
}

// Join is strings.Join, but variadic.
func Join(prefix string, xs ...string) string { return strings.Join(xs, prefix) }

// CopyStringMap returns a new copy of a map of strings.
func CopyStringMap(xs map[string]string) map[string]string {
	// Zero allocation path.
	if xs == nil {
		return nil
	}

	ys := make(map[string]string, len(xs))
	for k, v := range xs {
		ys[k] = v
	}
	return ys
}

// Unpack assigns a slice into local variables.
func Unpack(xs []string, vars ...*string) {
	for i, x := range xs {
		*vars[i] = x
	}
}
