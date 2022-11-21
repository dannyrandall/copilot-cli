// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type unionTest[A, B any] struct {
	yaml string

	expectedValue        Union[A, B]
	expectedUnmarshalErr string
	expectedYAML         string
}

func TestUnion(t *testing.T) {
	runUnionTest(t, "string or []string, is string", unionTest[string, []string]{
		yaml: `key: hello`,
		expectedValue: Union[string, []string]{
			Basic: "hello",
		},
	})
	runUnionTest(t, "string or []string, is zero string", unionTest[string, []string]{
		yaml: `key: ""`,
	})
	runUnionTest(t, "string or []string, is []string", unionTest[string, []string]{
		yaml: `
key:
  - asdf
  - jkl;`,
		expectedValue: Union[string, []string]{
			Advanced: []string{"asdf", "jkl;"},
		},
	})
	runUnionTest(t, "*string or []string, is string", unionTest[*string, []string]{
		yaml: `key: hello`,
		expectedValue: Union[*string, []string]{
			Basic: aws.String("hello"),
		},
	})
	runUnionTest(t, "*string or []string, is zero string", unionTest[*string, []string]{
		yaml: `key: ""`,
		expectedValue: Union[*string, []string]{
			Basic: aws.String(""),
		},
	})
	runUnionTest(t, "*string or []string, is null", unionTest[*string, semiComplexStruct]{
		yaml: `key: null`,
	})
	runUnionTest(t, "bool or semiComplexStruct, is false bool", unionTest[bool, semiComplexStruct]{
		yaml: `key: false`,
	})
	runUnionTest(t, "bool or semiComplexStruct, is true bool", unionTest[bool, semiComplexStruct]{
		yaml: `key: true`,
		expectedValue: Union[bool, semiComplexStruct]{
			Basic: true,
		},
	})
	runUnionTest(t, "*bool or semiComplexStruct, is false bool", unionTest[*bool, semiComplexStruct]{
		yaml: `key: false`,
		expectedValue: Union[*bool, semiComplexStruct]{
			Basic: aws.Bool(false),
		},
	})
	runUnionTest(t, "*bool or semiComplexStruct, is null", unionTest[*bool, semiComplexStruct]{
		yaml: `key: null`,
	})
	runUnionTest(t, "bool or semiComplexStruct, is semiComplexStruct with all fields set", unionTest[bool, semiComplexStruct]{
		yaml: `
key:
  str: asdf
  bool: true
  int: 420
  str_ptr: jkl;
  bool_ptr: false
  int_ptr: 70`,
		expectedValue: Union[bool, semiComplexStruct]{
			Advanced: semiComplexStruct{
				Str:     "asdf",
				Bool:    true,
				Int:     420,
				StrPtr:  aws.String("jkl;"),
				BoolPtr: aws.Bool(false),
				IntPtr:  aws.Int(70),
			},
		},
	})
	runUnionTest(t, "bool or semiComplexStruct, is semiComplexStruct without strs set", unionTest[bool, semiComplexStruct]{
		yaml: `
key:
  bool: true
  int: 420
  bool_ptr: false
  int_ptr: 70`,
		expectedValue: Union[bool, semiComplexStruct]{
			Advanced: semiComplexStruct{
				Bool:    true,
				Int:     420,
				BoolPtr: aws.Bool(false),
				IntPtr:  aws.Int(70),
			},
		},
	})
	runUnionTest(t, "string or semiComplexStruct, is struct with invalid fields", unionTest[string, semiComplexStruct]{
		// This case is useful to demonstrate the effects of the Unmarshal() logic.
		// Because there is no error unmarshalling into the struct (without strict mode)
		// this results in the value of the Union being the zero value of Basic (a string).
		yaml: `
key:
  invalid_key: asdf`,
		expectedYAML: `key: ""`,
	})
	runUnionTest(t, "semiComplexStruct or complexStruct, is complexStruct with all fields", unionTest[semiComplexStruct, complexStruct]{
		// This case shows that when both types are successfully unmarshalled into,
		// both types are set, but Marshal() considers the value as Advanced.
		yaml: `
key:
  str_ptr: qwerty
  semi_complex_struct:
    str: asdf
    bool: true
    int: 420
    str_ptr: jkl;
    bool_ptr: false
    int_ptr: 70`,
		expectedValue: Union[semiComplexStruct, complexStruct]{
			Basic: semiComplexStruct{
				StrPtr: aws.String("qwerty"),
			},
			Advanced: complexStruct{
				StrPtr: aws.String("qwerty"),
				SemiComplexStruct: semiComplexStruct{
					Str:     "asdf",
					Bool:    true,
					Int:     420,
					StrPtr:  aws.String("jkl;"),
					BoolPtr: aws.Bool(false),
					IntPtr:  aws.Int(70),
				},
			},
		},
	})
	runUnionTest(t, "string or bool, is []string, error", unionTest[string, bool]{
		yaml: `
key:
  - asdf`,
		expectedUnmarshalErr: `unmarshal to basic form string: yaml: unmarshal errors:
  line 3: cannot unmarshal !!seq into string
unmarshal to advanced form bool: yaml: unmarshal errors:
  line 3: cannot unmarshal !!seq into bool`,
		expectedYAML: `key: ""`,
	})
	runUnionTest(t, "*bool or *string, is []string, error", unionTest[*bool, *string]{
		yaml: `

key:
  - asdf`,
		expectedUnmarshalErr: `unmarshal to basic form *bool: yaml: unmarshal errors:
  line 4: cannot unmarshal !!seq into bool
unmarshal to advanced form *string: yaml: unmarshal errors:
  line 4: cannot unmarshal !!seq into string`,
		expectedYAML: `key: null`,
	})
	runUnionTest(t, "[]string or semiComplexStruct, is []string", unionTest[[]string, semiComplexStruct]{
		yaml: `
key:
  - asdf`,
		expectedValue: Union[[]string, semiComplexStruct]{
			Basic: []string{"asdf"},
		},
	})
	runUnionTest(t, "[]string or semiComplexStruct, is semiComplexStruct", unionTest[[]string, semiComplexStruct]{
		yaml: `
key:
  bool: true
  int: 420`,
		expectedValue: Union[[]string, semiComplexStruct]{
			Advanced: semiComplexStruct{
				Bool: true,
				Int:  420,
			},
		},
	})
	runUnionTest(t, "[]string or semiComplexStruct, is string, error", unionTest[[]string, semiComplexStruct]{
		yaml:                 `key: asdf`,
		expectedUnmarshalErr: "unmarshal to basic form []string: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `asdf` into []string\nunmarshal to advanced form manifest.semiComplexStruct: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `asdf` into manifest.semiComplexStruct",
		expectedYAML:         `key: []`,
	})
	runUnionTest(t, "string or semiComplexStruct, never instantiated", unionTest[string, semiComplexStruct]{
		yaml:         `wrongkey: asdf`,
		expectedYAML: `key: ""`,
	})
	runUnionTest(t, "semiComplexStruct or isZeroer, is non-zero isZeroer", unionTest[semiComplexStruct, isZeroer]{
		yaml: `
key:
  subkey: "asdf"`,
		expectedValue: Union[semiComplexStruct, isZeroer]{
			Advanced: isZeroer{
				SubKey: "asdf",
			},
		},
		expectedYAML: `
key:
  subkey: asdf`,
	})
	runUnionTest(t, "semiComplexStruct or isZeroer, is zero isZeroer", unionTest[semiComplexStruct, isZeroer]{
		yaml: `
key:
  subkey: "iamzero"`,
		expectedValue: Union[semiComplexStruct, isZeroer]{
			Advanced: isZeroer{
				SubKey: "iamzero",
			},
		},
		expectedYAML: `
key:
  bool: false
  int: 0`,
	})
}

type keyValue[Basic, Advanced any] struct {
	Key Union[Basic, Advanced] `yaml:"key"`
}

func runUnionTest[Basic, Advanced any](t *testing.T, name string, test unionTest[Basic, Advanced]) {
	t.Run(name, func(t *testing.T) {
		var kv keyValue[Basic, Advanced]
		dec := yaml.NewDecoder(strings.NewReader(test.yaml))

		err := dec.Decode(&kv)
		if test.expectedUnmarshalErr != "" {
			require.EqualError(t, err, test.expectedUnmarshalErr)
		} else {
			require.NoError(t, err)
		}

		require.Equal(t, test.expectedValue, kv.Key)

		// call Marshal() with an indent of 2 spaces
		buf := &bytes.Buffer{}
		enc := yaml.NewEncoder(buf)
		enc.SetIndent(2)
		err = enc.Encode(kv)
		require.NoError(t, err)
		require.NoError(t, enc.Close())

		expectedYAML := test.yaml
		if test.expectedYAML != "" {
			expectedYAML = test.expectedYAML
		}

		// verify the marshaled string matches the input string
		require.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(buf.String()))
	})
}

func TestUnion_EmbeddedType(t *testing.T) {
	type embeddedType struct {
		Union[string, []string]
	}

	type keyValue struct {
		Key embeddedType `yaml:"key,omitempty"`
	}

	// test []string
	in := `
key:
  - asdf
`
	var kv keyValue
	require.NoError(t, yaml.Unmarshal([]byte(in), &kv))
	require.Equal(t, keyValue{
		Key: embeddedType{
			Union[string, []string]{
				Advanced: []string{"asdf"},
			},
		},
	}, kv)

	// test string
	in = `
key: qwerty
`
	kv = keyValue{}
	require.NoError(t, yaml.Unmarshal([]byte(in), &kv))
	require.Equal(t, keyValue{
		Key: embeddedType{
			Union[string, []string]{
				Basic: "querty",
			},
		},
	}, kv)
}

type semiComplexStruct struct {
	Str     string  `yaml:"str,omitempty"`
	Bool    bool    `yaml:"bool"`
	Int     int     `yaml:"int"`
	StrPtr  *string `yaml:"str_ptr,omitempty"`
	BoolPtr *bool   `yaml:"bool_ptr,omitempty"`
	IntPtr  *int    `yaml:"int_ptr,omitempty"`
}

type complexStruct struct {
	StrPtr            *string           `yaml:"str_ptr,omitempty"`
	SemiComplexStruct semiComplexStruct `yaml:"semi_complex_struct"`
}

type isZeroer struct {
	SubKey string `yaml:"subkey"`
}

func (a isZeroer) IsZero() bool {
	return a.SubKey == "iamzero"
}
