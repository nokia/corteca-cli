// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package templating

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/spf13/afero"
)

func TestRenderTemplateFile(t *testing.T) {
	fs := afero.NewMemMapFs()

	//Mock template file
	templateContent := "Hello, {{.Name}}!"
	templateDir := "/project"
	templateFile := "template.txt"
	fs.MkdirAll(templateDir, 0755)
	afero.WriteFile(fs, filepath.Join(templateDir, templateFile), []byte(templateContent), 0644)

	testCases := []struct {
		name           string
		relativePath   string
		srcDir         string
		destRootDir    string
		context        any
		wantErr        bool
		wantOutputFile string
		wantOutput     string
	}{
		{
			name:           "ValidTemplate",
			relativePath:   templateFile,
			srcDir:         templateDir,
			destRootDir:    "/dest",
			context:        map[string]string{"Name": "Alice"},
			wantErr:        false,
			wantOutputFile: "/dest/template.txt",
			wantOutput:     "Hello, Alice!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RenderTemplateFile(fs, tc.relativePath, tc.srcDir, tc.destRootDir, tc.context)
			if (err != nil) != tc.wantErr {
				t.Errorf("RenderTemplateFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				content, readErr := afero.ReadFile(fs, tc.wantOutputFile)
				if readErr != nil {
					t.Fatalf("Failed to read output file: %v", readErr)
				}
				if string(content) != tc.wantOutput {
					t.Errorf("RenderTemplateFile() got output = %v, want %v", string(content), tc.wantOutput)
				}
			}
		})
	}
}

func TestHasEmptyElem(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"/path/to/dir/", true},
		{"/path//folder empty spaces/dir", true},
		{"", true},
		{"/path/to///dir", true},
		{"normal/path", false},
		{"/", false},
		{"/path/to/ /dir", true},
		{"/path/to/test app/dir", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := hasEmptyElem(tc.path)
			if result != tc.expected {
				t.Errorf("hasEmptyElem(%q) = %v; want %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestRenderTemplateString(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		context  map[string]any
		want     string
		err      bool
	}{
		{
			name:     "Replace single placeholder",
			template: "Hello, {{.Name}}!",
			context:  map[string]any{"Name": "Jane"},
			want:     "Hello, Jane!",
			err:      false,
		},
		{
			name:     "Replace multiple placeholders",
			template: "{{.Greeting}}, {{.Name}}!",
			context:  map[string]any{"Greeting": "Hi", "Name": "John"},
			want:     "Hi, John!",
			err:      false,
		},
		{
			name:     "No placeholders",
			template: "Hello, World!",
			context:  map[string]any{},
			want:     "Hello, World!",
			err:      false,
		},
		{
			name:     "Missing context value",
			template: "Hello, {{.Name}} and {{.Friend}}!",
			context:  map[string]any{"Name": "Jane"},
			want:     "",
			err:      true,
		},
		{
			name:     "Special characters in template",
			template: "Values: {{.Value1}}, {{.Value2}}, and {{.Value3}}!",
			context:  map[string]any{"Value1": "@", "Value2": "%", "Value3": "#"},
			want:     "Values: @, %, and #!",
			err:      false,
		},
		{
			name:     "Complex context values",
			template: "User: {{.User.Name}}, Age: {{.User.Age}}",
			context:  map[string]any{"User": map[string]any{"Name": "Jane", "Age": 30}},
			want:     "User: Jane, Age: 30",
			err:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RenderTemplateString(tc.template, tc.context)
			if tc.err && (err == nil) {
				t.Errorf("%v: expected error, nil received", tc.name)
				return
			} else if !tc.err && (err != nil) {
				t.Errorf("%v: %v", tc.name, err.Error())
				return
			}
			if got != tc.want {
				t.Errorf("%v: received %q\nexpected %q", t.Name(), got, tc.want)
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
