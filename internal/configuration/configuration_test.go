// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package configuration

import (
	"bytes"
	"reflect"
	"sort"
	"testing"

	"github.com/spf13/afero"
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
		wantOutput any
	}{
		{
			name:       "ReadTopLevelField",
			fieldPath:  "app.lang",
			wantErr:    false,
			wantOutput: "go",
		},
		{
			name:       "ReadMapField",
			fieldPath:  "app.options",
			wantErr:    false,
			wantOutput: map[string]any{"option1": "value1"},
		},
		{
			name:      "InvalidFieldPath",
			fieldPath: "app.nonExistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := conf.ReadField(tt.fieldPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(value, tt.wantOutput) {
				t.Errorf("ReadField() got output = %v, want %v", value, tt.wantOutput)
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
			wantFound: true,
			wantValue: "value2",
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
			gotField := fieldByEncodingName(tt.v, tt.yamlName)
			if gotField.IsValid() != tt.wantFound {
				t.Errorf("fieldByEncodingName() for %v, found = %v, want %v", tt.yamlName, gotField.IsValid(), tt.wantFound)
			} else if tt.wantFound && gotField.String() != tt.wantValue {
				t.Errorf("fieldByEncodingName() got = %v, want %v", gotField.String(), tt.wantValue)
			}
		})
	}
}

func TestWriteField(t *testing.T) {
	// Mock Settings object
	newConf := func() *Settings {
		return &Settings{
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
			Publish: map[string]PublishTarget{
				"local": {
					Endpoint: Endpoint{
						Addr: "http://0.0.0.0:8080",
					},
				},
			},
		}
	}

	tests := []struct {
		name      string
		fieldPath string
		wantErr   bool
		value     string
		expected  string
		append    bool
		actual    func(c *Settings) any
	}{
		{
			name:      "TestWritePlainStructField",
			fieldPath: "app.lang",
			value:     "python",
			actual:    func(c *Settings) any { return c.App.Lang },
		},
		{
			name:      "TestWritePlainMapField",
			fieldPath: "publish.local.addr",
			value:     "172.17.0.1:9000",
			actual:    func(c *Settings) any { return c.Publish["local"].Addr },
		},
		{
			name:      "TestWriteWrongStructField",
			fieldPath: "app.foo",
			value:     "",
			wantErr:   true,
		},
		{
			name:      "TestWriteWrongMapField",
			fieldPath: "app.env.foo",
			value:     "",
			wantErr:   true,
		},
		{
			name:      "TestWriteComplexStructField",
			fieldPath: "app.options",
			value:     "{foo: {bar: zed}}",
			actual:    func(c *Settings) any { return c.App.Options },
		},
		{
			name:      "TestAppendSliceField",
			fieldPath: "app.dependencies.compile",
			value:     "foobar",
			append:    true,
			actual:    func(c *Settings) any { return c.App.Dependencies.Compile[len(c.App.Dependencies.Compile)-1] },
		},
		{
			name:      "TestAppendMapField",
			fieldPath: "app.env",
			value:     "{foobar: zed}",
			expected:  "zed",
			append:    true,
			actual:    func(c *Settings) any { return c.App.Env["foobar"] },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := newConf()
			err := conf.WriteField(tt.fieldPath, tt.value, tt.append)

			if (err != nil) != tt.wantErr {
				t.Errorf("%s failed with error: %s", tt.name, err.Error())
				return
			}

			if !tt.wantErr {
				actualVal, _ := yaml.Marshal(tt.actual(conf))

				// unmarshal/marshal value to avoid formatting differences
				var vn any
				if len(tt.expected) == 0 {
					yaml.Unmarshal([]byte(tt.value), &vn)
				} else {
					yaml.Unmarshal([]byte(tt.expected), &vn)
				}
				expectedVal, _ := yaml.Marshal(vn)

				if !bytes.Equal(actualVal, expectedVal) {
					t.Errorf("%s fail with wrong output\n\tactual: '%s'\n\texpected: '%s'", tt.name, actualVal, expectedVal)
				}
			}
		})
	}
}

func TestGetSuggestions(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		suggestions []string
	}{
		{
			path:        "app.version.",
			suggestions: nil,
		},
		{
			path:        "app.version",
			suggestions: []string{"app.version"},
		},
		{
			path: "build.",
			suggestions: []string{
				"build.crossCompile",
				"build.default",
				"build.dockerFileTemplate",
				"build.options",
				"build.toolchains",
			},
		},
		{
			path: "build.d",
			suggestions: []string{
				"build.default",
				"build.dockerFileTemplate",
			},
		},
		{
			path: "app.dependencies.compile.",
			suggestions: []string{
				"app.dependencies.compile.0",
				"app.dependencies.compile.1",
				"app.dependencies.compile.2",
			},
		},
		{
			path: "app.env.",
			suggestions: []string{
				"app.env.baz",
				"app.env.foo",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewConfiguration()
			config.App.Dependencies.Compile = append(config.App.Dependencies.Compile, "foo", "bar", "baz")
			config.App.Env = map[string]string{
				"foo": "bar",
				"baz": "zed",
			}
			suggestions := config.GetSuggestions(tc.path)
			if !reflect.DeepEqual(suggestions, tc.suggestions) {
				t.Errorf("%s: failed with path '%s'.\n\texpected: %v\n\tactual: %v\n", t.Name(), tc.path, tc.suggestions, suggestions)
			}
		})
	}
}

func TestGetFileList(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Mock file structure
	afero.WriteFile(fs, "/project/file1.txt", []byte("file1"), 0644)
	afero.WriteFile(fs, "/project/file2.txt", []byte("file2"), 0644)
	fs.MkdirAll("/project/subdir", 0755)
	afero.WriteFile(fs, "/project/subdir/file3.txt", []byte("file3"), 0644)
	fs.MkdirAll("/empty", 0755)
	testCases := []struct {
		name       string
		rootFolder string
		want       []string
		wantErr    bool
	}{
		{
			name:       "DirectoryWithFiles",
			rootFolder: "/project",
			want:       []string{"file1.txt", "file2.txt", "subdir/file3.txt"},
			wantErr:    false,
		},
		{
			name:       "EmptyDirectory",
			rootFolder: "/empty",
			want:       nil,
			wantErr:    false,
		},
		{
			name:       "NonExistentDirectory",
			rootFolder: "/nonexistent",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getFileList(fs, tc.rootFolder)
			if (err != nil) != tc.wantErr {
				t.Errorf("getFileList() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			sort.Strings(got) // Sort the slices for consistent comparison
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getFileList() got = %v, want %v", got, tc.want)
			}
		})
	}
}
