// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

// Package templating handles rendering folder structures with templated names
// and file content, based on go's text/template package
package templating

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"
)

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

	err = RenderTemplateToFile(fs, templateFilePath, outputFile, context)
	if err != nil {
		return err
	}

	return nil
}

func GetEnvVar(envVarName string, defaultValue ...string) string {
	envVarValue, exists := os.LookupEnv(envVarName)
	if !exists && len(defaultValue) == 0 {
		return defaultValue[0]
	}
	return envVarValue
}

func RenderTemplateToFile(fs afero.Fs, srcFile, destFile string, context any) error {
	templateContent, err := afero.ReadFile(fs, srcFile)
	if err != nil {
		return err
	}

	fileTmpl, err := template.New(filepath.Base(srcFile)).Option("missingkey=error").Parse(string(templateContent))

	if err != nil {
		return err
	}

	destDir := filepath.Dir(destFile)
	if err := fs.MkdirAll(destDir, 0777); err != nil {
		return err
	}

	out, err := fs.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := fileTmpl.Execute(out, context); err != nil {
		return err
	}

	srcFileInfo, err := fs.Stat(srcFile)
	if err != nil {
		return err
	}
	if err := fs.Chmod(destFile, srcFileInfo.Mode()); err != nil {
		return err
	}

	return nil
}
