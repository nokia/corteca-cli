// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

// Package templating handles rendering folder structures with templated names
// and file content, based on go's text/template package
package templating

import (
	"bytes"
	"corteca/internal/configuration"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"
)

const TemplateInfoFile string = ".template-info.yaml"

type TemplateCustomOption struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Default     any      `yaml:"default"`
	Values      []string `yaml:"values,omitempty"`
}

type TemplateInfo struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Dependencies struct {
		Compile []string `yaml:"compile"`
		Runtime []string `yaml:"runtime"`
	} `yaml:"dependencies"`
	Path    string `yaml:"-"`
	Options []TemplateCustomOption
}

const (
	BoolOption   = "boolean"
	TextOption   = "text"
	ChoiceOption = "choice"
)

// return a map (template name) ->  TemplateInfo of available templates
func GetAvailableTemplates(list map[string]TemplateInfo, templatesDir string) error {
	// find all files in folder (recursively) that match the template info filename
	err := fs.WalkDir(os.DirFS(templatesDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		filename := filepath.Base(path)
		if filename == TemplateInfoFile {
			fullPath := filepath.Join(templatesDir, path)
			// index template name from parent folder name
			info, err := readTemplateInfo(fullPath)
			if err != nil {
				return err
			}
			list[filepath.Base(info.Path)] = info
		}
		return nil
	})
	// consume error in case folder does not exist
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return nil
}

func GenerateTemplate(fs afero.Fs, info TemplateInfo, destFolder string, context any) error {
	fileList, err := getFileList(fs, info.Path)
	if err != nil {
		return err
	}
	for _, path := range fileList {
		if filepath.Base(path) == TemplateInfoFile {
			continue
		}
		err := RenderTemplateFile(fs, path, info.Path, destFolder, context)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFileList(fs afero.Fs, rootFolder string) ([]string, error) {
	var fileList []string
	err := afero.Walk(afero.NewBasePathFs(fs, rootFolder), ".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}

func RenderTemplateString(tmpl string, context any) (string, error) {
	nameTmpl, err := template.New("").Funcs(template.FuncMap{"getEnv": GetEnvVar}).Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var renderedName bytes.Buffer
	// in case of error, return unchanged template
	if err = nameTmpl.Execute(&renderedName, context); err != nil {
		return "", err
	}
	return renderedName.String(), nil
}

func hasEmptyElem(filepath string) bool {

	// Edge case: root directory and empty string
	if filepath == "/" {
		return false
	}

	// Edge case: empty string
	if filepath == "" {
		return true
	}

	filepath = strings.TrimLeft(filepath, string(os.PathSeparator))
	parts := strings.Split(filepath, string(os.PathSeparator))

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return true
		}
	}

	return false
}

func RenderTemplateFile(fs afero.Fs, relativePath, srcDir, destRootDir string, context any) error {
	outputFile, err := RenderTemplateString(relativePath, context)
	if err != nil {
		return err
	}
	if hasEmptyElem(outputFile) {
		return nil
	}
	outputFile = filepath.Join(destRootDir, outputFile)

	templateFilePath := filepath.Join(srcDir, relativePath)

	// Read the template file using Afero
	templateContent, err := afero.ReadFile(fs, templateFilePath)
	if err != nil {
		return err
	}

	// Parse the template content
	fileTmpl, err := template.New(filepath.Base(relativePath)).Option("missingkey=error").Parse(string(templateContent))
	if err != nil {
		return err
	}

	destDir := filepath.Dir(outputFile)
	if err := fs.MkdirAll(destDir, 0777); err != nil {
		return err
	}

	out, err := fs.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := fileTmpl.Execute(out, context); err != nil {
		return err
	}

	return nil
}

func readTemplateInfo(fullPath string) (info TemplateInfo, err error) {
	yamlData, err := os.Open(fullPath)
	if err != nil {
		return TemplateInfo{}, err
	}
	defer yamlData.Close()

	if err = configuration.ReadYamlInto(&info, yamlData); err != nil {
		return
	}
	// TODO: validate template info
	info.Path = filepath.Dir(fullPath)
	return
}

func GetEnvVar(envVarName string, defaultValue ...string) string {
	envVarValue, exists := os.LookupEnv(envVarName)
	if !exists && len(defaultValue) == 0 {
		return defaultValue[0]
	}
	return envVarValue
}
