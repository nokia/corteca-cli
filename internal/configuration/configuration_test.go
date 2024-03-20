// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package configuration

import (
	"bytes"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// Mock structs
type InnerStruct struct {
	FieldA string `yaml:"fieldA"`
}

type TestStruct struct {
	SimpleField string            `yaml:"simpleField"`
	NestedField InnerStruct       `yaml:"nestedField"`
	MapField    map[string]string `yaml:"mapField"`
}

type TestStructWithTags struct {
	FieldWithYamlTag    string `yaml:"fieldYaml"`
	FieldWithoutYamlTag string
	FieldWithOmitempty  string `yaml:"omitemptyField,omitempty"`
}

func TestReadField(t *testing.T) {
	// Mock Settings object
	conf := &Settings{
		App: AppSettings{
			Lang:        "go",
			Title:       "My App",
			Name:        "my_app",
			Author:      "Jane Doe",
			Description: "A sample app",
			Version:     "1.0.0",
			FQDN:        "domain.example.com",
			DUID:        "50e02eca-5e37-5365-8d95-9ec69e2512f7",
			Options: map[string]any{
				"option1": "value1",
			},
		},
	}

	tests := []struct {
		name       string
		fieldPath  string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "ReadTopLevelField",
			fieldPath:  "app.lang",
			wantErr:    false,
			wantOutput: "go\n",
		},
		{
			name:       "ReadMapField",
			fieldPath:  "app.options",
			wantErr:    false,
			wantOutput: "option1: value1\n",
		},
		{
			name:      "InvalidFieldPath",
			fieldPath: "app.nonExistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := conf.ReadField(tt.fieldPath, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && buf.String() != tt.wantOutput {
				t.Errorf("ReadField() got output = %v, want %v", buf.String(), tt.wantOutput)
			}
		})
	}
}

func TestSettingsToDictionary(t *testing.T) {
	type TestCaseContextSub struct {
		SubFieldA string `yaml:"foo"`
		SubFieldB int    `yaml:"SubFieldB"`
	}
	type TestCaseContext struct {
		FieldA string             `yaml:"fieldA"`
		FieldB int                `yaml:"fieldB"`
		FieldC TestCaseContextSub `yaml:"fieldC"`
	}

	testCases := []struct {
		name    string
		context TestCaseContext
		want    map[string]any
		wantErr bool
	}{
		{
			name: "ValidSettings",
			context: TestCaseContext{
				FieldA: "a value",
				FieldB: 42,
				FieldC: TestCaseContextSub{
					SubFieldA: "another value",
					SubFieldB: 0xdeadbeef,
				},
			},
			want: map[string]any{
				"fieldA": "a value",
				"fieldB": 42,
				"fieldC": map[string]any{
					"foo":       "another value",
					"SubFieldB": 0xdeadbeef,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotDict := ToDictionary(tt.context)

			// Convert both gotDict and tt.want to YAML strings for comparison
			gotYAML, _ := yaml.Marshal(gotDict)
			wantYAML, _ := yaml.Marshal(tt.want)

			if string(gotYAML) != string(wantYAML) {
				t.Errorf("%v:\nreceived: %v\nexpected: %v", t.Name(), string(gotYAML), string(wantYAML))
			}
		})
	}

}

func TestComputeDelta(t *testing.T) {

	testCases := []struct {
		name     string
		prev     map[string]any
		curr     map[string]any
		expected map[string]any
	}{
		{
			name:     "IdenticalMaps",
			prev:     map[string]any{"key1": "value1", "key2": "value2"},
			curr:     map[string]any{"key1": "value1", "key2": "value2"},
			expected: map[string]any{},
		},
		{
			name:     "DifferentValues",
			prev:     map[string]any{"key1": "value1"},
			curr:     map[string]any{"key1": "newValue"},
			expected: map[string]any{"key1": "newValue"},
		},
		{
			name:     "SubsetMap",
			prev:     map[string]any{"key1": "value1", "key2": "value2"},
			curr:     map[string]any{"key2": "value2"},
			expected: map[string]any{},
		},
		{
			name:     "NestedMapWithDifferences",
			prev:     map[string]any{"key1": map[string]any{"nestedKey1": "nestedValue1"}},
			curr:     map[string]any{"key1": map[string]any{"nestedKey1": "nestedValue2"}},
			expected: map[string]any{"key1": map[string]any{"nestedKey1": "nestedValue2"}},
		},
		{
			name:     "MixedValueTypes",
			prev:     map[string]any{"key1": "value1", "key2": 100},
			curr:     map[string]any{"key1": "value1", "key2": 200},
			expected: map[string]any{"key2": 200},
		},
		{
			name:     "EmptyCurrentMap",
			prev:     map[string]any{"key1": "value1", "key2": "value2"},
			curr:     map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "NestedMapWithAdditionalKeys",
			prev:     map[string]any{"key1": map[string]any{"nestedKey1": "nestedValue1"}},
			curr:     map[string]any{"key1": map[string]any{"nestedKey1": "nestedValue1", "nestedKey2": "nestedValue2"}},
			expected: map[string]any{"key1": map[string]any{"nestedKey2": "nestedValue2"}},
		},
		{
			name:     "ValueChangedToMap",
			prev:     map[string]any{"key1": "value1"},
			curr:     map[string]any{"key1": map[string]any{"nestedKey": "nestedValue"}},
			expected: map[string]any{"key1": map[string]any{"nestedKey": "nestedValue"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := computeDelta(tc.prev, tc.curr)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test %s failed. Expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}

func TestRetrieveField(t *testing.T) {
	testObj := TestStruct{
		SimpleField: "Value",
		NestedField: InnerStruct{FieldA: "NestedValue"},
		MapField:    map[string]string{"key": "mapValue"},
	}

	tests := []struct {
		name        string
		v           reflect.Value
		fieldPath   string
		addressable bool
		wantErr     bool
		wantValue   interface{}
	}{
		{
			name:        "TopLevelField",
			v:           reflect.ValueOf(&testObj),
			fieldPath:   "simpleField",
			addressable: false,
			wantErr:     false,
			wantValue:   "Value",
		},
		{
			name:        "NestedField",
			v:           reflect.ValueOf(&testObj),
			fieldPath:   "nestedField.fieldA",
			addressable: false,
			wantErr:     false,
			wantValue:   "NestedValue",
		},
		{
			name:        "MapField",
			v:           reflect.ValueOf(&testObj),
			fieldPath:   "mapField",
			addressable: true,
			wantErr:     false,
			wantValue:   testObj.MapField,
		},
		{
			name:        "InvalidField",
			v:           reflect.ValueOf(&testObj),
			fieldPath:   "NonExistentField",
			addressable: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotField, _, err := retrieveField(tt.v, tt.fieldPath, tt.addressable)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotField.Kind() == reflect.Ptr {
					gotField = gotField.Elem()
				}
				if !reflect.DeepEqual(gotField.Interface(), tt.wantValue) {
					t.Errorf("retrieveField() got = %v, want %v", gotField.Interface(), tt.wantValue)
				}
			}
		})
	}
}

func TestFieldByEncodingName(t *testing.T) {
	testStruct := TestStructWithTags{
		FieldWithYamlTag:    "value1",
		FieldWithoutYamlTag: "value2",
		FieldWithOmitempty:  "value3",
	}

	tests := []struct {
		name      string
		v         reflect.Value
		yamlName  string
		wantFound bool
		wantValue string
	}{
		{
			name:      "FieldWithYamlTag",
			v:         reflect.ValueOf(testStruct),
			yamlName:  "fieldYaml",
			wantFound: true,
			wantValue: "value1",
		},
		{
			name:      "FieldWithoutYamlTag",
			v:         reflect.ValueOf(testStruct),
			yamlName:  "FieldWithoutYamlTag",
			wantFound: false,
		},
		{
			name:      "FieldWithOmitempty",
			v:         reflect.ValueOf(testStruct),
			yamlName:  "omitemptyField",
			wantFound: true,
			wantValue: "value3",
		},
		{
			name:      "NonExistentField",
			v:         reflect.ValueOf(testStruct),
			yamlName:  "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotField := fieldByEncodingName(tt.v, tt.yamlName, "yaml")
			if gotField.IsValid() != tt.wantFound {
				t.Errorf("fieldByEncodingName() for %v, found = %v, want %v", tt.yamlName, gotField.IsValid(), tt.wantFound)
			} else if tt.wantFound && gotField.String() != tt.wantValue {
				t.Errorf("fieldByEncodingName() got = %v, want %v", gotField.String(), tt.wantValue)
			}
		})
	}
}

func TestSetKeyValuePair(t *testing.T) {
	testCases := []struct {
		name      string
		initMap   map[string]any
		fieldPath string
		value     any
		want      map[string]any
	}{
		{
			name:      "SetTopLevelKey",
			initMap:   map[string]any{},
			fieldPath: "key1",
			value:     "value1",
			want:      map[string]any{"key1": "value1"},
		},
		{
			name:      "SetNestedKey",
			initMap:   map[string]any{},
			fieldPath: "nested.key2",
			value:     "value2",
			want:      map[string]any{"nested": map[string]any{"key2": "value2"}},
		},
		{
			name:      "OverwriteExistingValue",
			initMap:   map[string]any{"key1": "oldValue"},
			fieldPath: "key1",
			value:     "newValue",
			want:      map[string]any{"key1": "newValue"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setKeyValuePair(tc.initMap, tc.fieldPath, tc.value)
			if err != nil {
				t.Errorf("setKeyValuePair() error = %v", err)
			}
			if !reflect.DeepEqual(tc.initMap, tc.want) {
				t.Errorf("setKeyValuePair() got = %v, want %v", tc.initMap, tc.want)
			}
		})
	}
}
