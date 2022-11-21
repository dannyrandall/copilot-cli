// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Union is a type used for yaml keys that may be of type Basic or Advanced.
// Functions on Union consider the current value of Union
// to be Advanced if it is non-zero, determined by
// Advanced.IsZero() or reflect.IsZero(). If Advanced is zero,
// the current value is Basic. This means the "zero" value of Union,
// if neither value is set, is the zero value of type Basic.
type Union[Basic, Advanced any] struct {
	// Basic holds a potential value of Union. It is considered the
	// value of Union if Advanced is zero.
	Basic Basic

	// Advanced holds a potential value of Union. It is considered
	// the value of Union if it is non-zero.
	Advanced Advanced
}

// BasicToUnion creates a new Union[Basic, Advanced] with the underlying
// type set to Basic, holding val.
func BasicToUnion[Basic, Advanced any](val Basic) Union[Basic, Advanced] {
	return Union[Basic, Advanced]{
		Basic: val,
	}
}

// AdvancedToUnion creates a new Union[Basic, Advanced] with the underlying
// type set to Advanced, holding val.
func AdvancedToUnion[Basic, Advanced any](val Advanced) Union[Basic, Advanced] {
	return Union[Basic, Advanced]{
		Advanced: val,
	}
}

// UnmarshalYAML decodes value into types Basic or Advanced.
// If decoding does not return an error for a type, it is set on t.
// An error is returned if value returns an error while decoding into both types.
func (t *Union[Basic, Advanced]) UnmarshalYAML(value *yaml.Node) error {
	// reset to zero
	var basic Basic
	var advanced Advanced
	t.Basic = basic
	t.Advanced = advanced

	bErr := value.Decode(&basic)
	aErr := value.Decode(&advanced)
	if bErr != nil && aErr != nil {
		// multiline error because yaml.TypeError (which this likely is)
		// is already a multiline error
		return fmt.Errorf("unmarshal to basic form %T: %s\nunmarshal to advanced form %T: %s", t.Basic, bErr, t.Advanced, aErr)
	}

	if bErr == nil {
		t.Basic = basic
	}
	if aErr == nil {
		t.Advanced = advanced
	}
	return nil
}

// isZero returns true if:
//   - v is a yaml.Zeroer and IsZero().
//   - v is not a yaml.Zeroer and determined to be zero via reflection.
func isZero(v any) bool {
	if z, ok := v.(yaml.IsZeroer); ok {
		return z.IsZero()
	}
	return reflect.ValueOf(v).IsZero()
}

// MarshalYAML implements yaml.Marshaler.
func (t Union[_, _]) MarshalYAML() (interface{}, error) {
	if !isZero(t.Advanced) {
		return t.Advanced, nil
	}
	return t.Basic, nil
}

// IsZero returns true if the set value of t is determined to be zero
// via yaml.Zeroer or reflection.
func (t Union[_, _]) IsZero() bool {
	if !isZero(t.Advanced) {
		return false
	}
	return isZero(t.Basic)
}

// validate calls t.validate() on the value of t. If the
// current value doesn't have a validate() function, it returns nil.
func (t Union[_, _]) validate() error {
	// type declarations inside generic functions not currently supported,
	// so we use an inline validate() interface
	if !isZero(t.Advanced) {
		if v, ok := any(t.Advanced).(interface{ validate() error }); ok {
			return v.validate()
		}
		return nil
	}

	if v, ok := any(t.Basic).(interface{ validate() error }); ok {
		return v.validate()
	}
	return nil
}

// SetBasic changes the value of the Union to v.
func (t *Union[Basic, Advanced]) SetBasic(v Basic) {
	var zero Advanced
	t.Advanced = zero
	t.Basic = v
}

// SetAdvanced changes the value of the Union to v.
func (t *Union[Basic, Advanced]) SetAdvanced(v Advanced) {
	var zero Basic
	t.Basic = zero
	t.Advanced = v
}
