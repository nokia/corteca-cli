// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

// Package config handles reading/writing configuration from/to a yaml file
//
// to implement:
// - read from multiple files and perform cascading
// - support multiple config contexts (corteca config and app config (manifest))
package configuration

import (
	specs "corteca/internal/configuration/runtimeSpec"
	"corteca/internal/configuration/templating"
	"corteca/internal/fsutil"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const (
	ENCODING            = "yaml"
	STRICT_DECODING_OPT = true
	INDENTATION         = 4
)

type TemplateField struct {
	RawTemplate string `yaml:"rawTemplate"`
}

type BuildOptions struct {
	OutputType  string            `yaml:"outputType"`
	DebugMode   bool              `yaml:"debug"`
	SkipHostEnv bool              `yaml:"skipHostEnv,omitempty"`
	Env         map[string]string `yaml:"env"`
}

type AppSettings struct {
	Name         string            `yaml:"name"`
	Author       string            `yaml:"author"`
	Version      string            `yaml:"version"`
	DUID         string            `yaml:"duid"`
	Dependencies Dependencies      `yaml:"dependencies"`
	Env          map[string]string `yaml:"env"`
	Entrypoint   []string          `yaml:"entrypoint"`
	Runtime      specs.Spec        `yaml:"runtime"`
}

type Dependencies struct {
	Compile []string `yaml:"compile"`
	Runtime []string `yaml:"runtime"`
}

type ArchitectureSettings struct {
	Platform string `yaml:"platform"`
}

type ArchitecturesMap map[string]ArchitectureSettings

type CrossCompileConfig struct {
	Enabled bool     `yaml:"enabled"`
	Image   string   `yaml:"image"`
	Args    []string `yaml:"args"`
}

type BuildSettings struct {
	Architectures ArchitecturesMap   `yaml:"architectures"`
	Default       string             `yaml:"default,omitempty"`
	Options       BuildOptions       `yaml:"options"`
	CrossCompile  CrossCompileConfig `yaml:"crossCompile"`
}

var CmdContext struct {
	App            *AppSettings      `yaml:"app,omitempty"`
	Arch           string            `yaml:"arch,omitempty"`
	BuildArtifacts map[string]string `yaml:"buildArtifacts,omitempty"`
	Device         struct {
		DeployDevice `yaml:",omitempty,inline"`
		Name         string `yaml:"name,omitempty"`
	} `yaml:"device,omitempty"`
	Publish struct {
		PublishTarget `yaml:",omitempty,inline"`
		Name          string `yaml:"name,omitempty"`
	} `yaml:"publish,omitempty"`
	Platform      string            `yaml:"platform,omitempty"`
	Build         *BuildSettings    `yaml:"build,omitempty"`
	BuildArtifact string            `yaml:"buildArtifact,omitempty"`
	Env           map[string]string `yaml:"env,omitempty"`
}

type PublishMethod int

type AuthType int

const (
	PUBLISH_METHOD_UNDEFINED = iota
	PUBLISH_METHOD_LISTEN
	PUBLISH_METHOD_PUT
	PUBLISH_METHOD_COPY
	PUBLISH_METHOD_PUSH
	PUBLISH_METHOD_REGISTRY
)

const (
	publishMethodListenName   = "listen"
	publishMethodPutName      = "put"
	publishMethodCopyName     = "copy"
	publishMethodPushName     = "push"
	publishMethodRegistryName = "registry-v2"
)

const (
	ConfigFileName = "corteca.yaml"
)

var cmdRegularExpression *regexp.Regexp
var regexKeyValue *regexp.Regexp
var exprRegex *regexp.Regexp

func init() {
	cmdRegularExpression = regexp.MustCompile(`^\s*\$\((.+)\)\s*$`)
	regexKeyValue = regexp.MustCompile(`^([[:word:]]+)=(.*)$`)
	exprRegex = regexp.MustCompile(`\${\s*(.*?)\s*}`)
	populateEnvVars()
}

func populateEnvVars() {
	CmdContext.Env = make(map[string]string)
	envVars := os.Environ()

	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			CmdContext.Env[parts[0]] = parts[1]
		}
	}
}

// encode TemplateField to YAML data
func (t TemplateField) MarshalYAML() (interface{}, error) {
	return t.RawTemplate, nil
}

// decode YAML data into TemplateField
func (t *TemplateField) UnmarshalYAML(data *yaml.Node) error {
	if data.Kind != yaml.ScalarNode {
		return errors.New("wrong value type")
	}

	t.RawTemplate = data.Value

	return nil
}

func evaluateExpressionFunc(visited []string, context any) func(string) string {
	if visited == nil {
		visited = make([]string, 0, 2)
	}
	return func(expr string) string {
		key := exprRegex.FindStringSubmatch(expr)[1]

		for i := range visited {
			if key == visited[i] {
				fmt.Printf("Warning: circular expression dependency detected for field: %s\n", key)
				return ""
			}
		}

		value, err := ReadField(context, key)

		if err != nil {
			fmt.Printf("Warning: could not read field '%s' with error: %v\n", key, err.Error())
			return ""
		}

		switch value.(type) {
		case TemplateField:
			visited = append(visited, key)
			return generateExpressions(value.(TemplateField).RawTemplate, visited, context)
		default:
			return fmt.Sprintf("%v", value)
		}
	}
}

func generateExpressions(input string, visited []string, context any) string {
	return exprRegex.ReplaceAllStringFunc(input, evaluateExpressionFunc(visited, context))
}

func (t TemplateField) String() string {
	return generateExpressions(t.RawTemplate, nil, CmdContext)
}

func (m PublishMethod) MarshalYAML() (interface{}, error) {
	var out []byte
	var err error

	switch m {
	case PUBLISH_METHOD_LISTEN:
		out, err = yaml.Marshal(publishMethodListenName)
	case PUBLISH_METHOD_PUT:
		out, err = yaml.Marshal(publishMethodPutName)
	case PUBLISH_METHOD_COPY:
		out, err = yaml.Marshal(publishMethodCopyName)
	case PUBLISH_METHOD_PUSH:
		out, err = yaml.Marshal(publishMethodPushName)
	case PUBLISH_METHOD_REGISTRY:
		out, err = yaml.Marshal(publishMethodRegistryName)
	default:
		out = nil
		err = fmt.Errorf("invalid publish method (%v)", m)
	}

	return strings.TrimSpace(string(out)), err

}

func (m *PublishMethod) UnmarshalYAML(data *yaml.Node) error {
	var name string
	if err := yaml.Unmarshal([]byte(data.Value), &name); err != nil {
		return err
	}
	name = strings.ToLower(name)
	switch name {
	case publishMethodListenName:
		*m = PUBLISH_METHOD_LISTEN
	case publishMethodPutName:
		*m = PUBLISH_METHOD_PUT
	case publishMethodCopyName:
		*m = PUBLISH_METHOD_COPY
	case publishMethodPushName:
		*m = PUBLISH_METHOD_PUSH
	case publishMethodRegistryName:
		*m = PUBLISH_METHOD_REGISTRY
	default:
		return fmt.Errorf("unrecognized publish method '%v'", name)
	}
	return nil
}

type Endpoint struct {
	Addr           TemplateField `yaml:"addr,omitempty"`
	Auth           string        `yaml:"auth,omitempty"`
	Username       TemplateField `yaml:"username,omitempty"`
	Password       TemplateField `yaml:"password,omitempty"`
	Password2      TemplateField `yaml:"password2,omitempty"`
	PrivateKeyFile TemplateField `yaml:"privateKeyFile,omitempty"`
	Token          TemplateField `yaml:"token,omitempty"`
}

type PublishTarget struct {
	Endpoint  `yaml:",omitempty,inline"`
	Method    PublishMethod `yaml:"method,omitempty"`
	PublicURL string        `yaml:"publicURL,omitempty"`
}

type SequenceCmd struct {
	Cmd           TemplateField `yaml:"cmd,omitempty"`
	Delay         uint          `yaml:"delay,omitempty"`
	Retries       uint          `yaml:"retries,omitempty"`
	Input         TemplateField `yaml:"input,omitempty"`
	IgnoreFailure bool          `yaml:"ignoreFailure,omitempty"`
}

type DownloadSource struct {
	Url     string `yaml:"url,omitempty"`
	Publish string `yaml:"publish,omitempty"`
}

type DeployDevice struct {
	Endpoint `yaml:",omitempty,inline"`
}

type Sequence []SequenceCmd

type DictType[T any] map[string]T

type Settings struct {
	App       AppSettings             `yaml:"app"`
	Build     BuildSettings           `yaml:"build"`
	Publish   DictType[PublishTarget] `yaml:"publish,omitempty"`
	Devices   DictType[DeployDevice]  `yaml:"devices,omitempty"`
	Sequences map[string]Sequence     `yaml:"sequences,omitempty"`
	Templates map[string]string       `yaml:"templates"`
}

// UnmarshalYAML for Publish and Devices
func (t *DictType[T]) UnmarshalYAML(data *yaml.Node) error {
	if data.Kind != yaml.MappingNode {
		return errors.New("wrong value type")
	}
	if *t == nil {
		*t = make(DictType[T])
	}
	for i := 0; i < len(data.Content); i += 2 {
		var alias string
		if err := data.Content[i].Decode(&alias); err != nil {
			return err
		}
		settings := (*t)[alias]
		if err := data.Content[i+1].Decode(&settings); err != nil {
			return err
		}

		(*t)[alias] = settings
	}
	return nil
}

type ExecuteCmdFunc func(string) error

func (c *Settings) ExecuteSequence(name string, context any, executeCmdFunc ExecuteCmdFunc) error {
	sequence, found := c.Sequences[name]
	if !found {
		return fmt.Errorf("sequence %s not found", name)
	}
	for idx, cmd := range sequence {
		fmt.Printf("Executing sequence '%s' step %d/%d...\n", name, idx+1, len(sequence))
		attempts := cmd.Retries + 1
		for {
			err := executeCommand(cmd, c, context, executeCmdFunc)
			attempts--
			if !cmd.IgnoreFailure && err != nil {
				if attempts == 0 {
					return err
				} else {
					fmt.Printf("Command failed (%s); will retry %d more time(s).\n", err.Error(), attempts)
				}
			}
			if cmd.Delay > 0 {
				fmt.Printf("=> Waiting for %d millisecond(s)...\n", cmd.Delay)
				time.Sleep(time.Duration(cmd.Delay) * time.Millisecond)
			}
			if err == nil {
				break
			}
		}
	}
	return nil
}

func findRefToSequence(seqCmd string) (string, bool) {
	if cmdRefRegex := cmdRegularExpression.FindStringSubmatch(seqCmd); len(cmdRefRegex) == 2 {
		return cmdRefRegex[1], true
	} else {
		return "", false
	}
}

func executeCommand(cmd SequenceCmd, c *Settings, context any, executeCmdFunc ExecuteCmdFunc) error {
	if seqName, found := findRefToSequence(cmd.Cmd.String()); found {
		return c.ExecuteSequence(seqName, context, executeCmdFunc)
	} else {
		cmdStr, err := templating.RenderTemplateString(cmd.Cmd.String(), context)
		if err != nil {
			if _, ok := err.(template.ExecError); ok {
				return fmt.Errorf("error rendering cmd content: %s", err.Error())
			}
			return err
		}
		fmt.Printf("=> Send cmd: '%s'...\n", cmdStr)
		return executeCmdFunc(cmdStr)
	}
}

func NewConfiguration() Settings {
	return Settings{
		App: AppSettings{
			Entrypoint: make([]string, 0),
		},
		Build: BuildSettings{
			Architectures: make(ArchitecturesMap),
		},
		Publish:   make(DictType[PublishTarget]),
		Devices:   make(DictType[DeployDevice]),
		Sequences: make(map[string]Sequence),
		Templates: make(map[string]string),
	}

}

func (conf *Settings) ReadFromFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	if err = ReadYamlInto(conf, in); err != nil {
		in.Close()
		return err
	}

	return nil
}

func (conf *Settings) WriteToFile(path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}

	enc := yaml.NewEncoder(out)
	enc.SetIndent(INDENTATION)

	if err = enc.Encode(conf); err != nil {
		return errors.Join(err, out.Close())
	}

	return out.Close()
}

const (
	INVALID_FIELD = "invalid field '%s'"
)

func (conf *Settings) GetSuggestions(fieldpath string) []string {
	keySequence := strings.Split(fieldpath, ".")
	field := reflect.ValueOf(*conf)
	// Preallocate slice with capacity 16 to minimize allocations for common cases
	suggestions := make([]string, 0, 16)

	for index, key := range keySequence {
		// if field is a pointer, dereference
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return nil
			}
			field = field.Elem()
		}

		if index == len(keySequence)-1 {
			// break on last item
			break
		}

		switch field.Kind() {
		case reflect.Struct:
			field = fieldByEncodingName(field, key)
			if !field.IsValid() {
				return nil
			}
		case reflect.Map:
			field = field.MapIndex(reflect.ValueOf(key))
			if !field.IsValid() {
				return nil
			}
		case reflect.Array, reflect.Slice:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= field.Len() {
				return nil
			}
			field = field.Index(i)
		default:
			return nil
		}
	}

	prefix := keySequence[len(keySequence)-1]
	path := strings.Join(keySequence[:len(keySequence)-1], ".")
	if path != "" {
		path += "."
	}

	switch field.Kind() {
	case reflect.Struct:
		suggestions = fieldNamesByEncodingPrefix(field, prefix, path)
	case reflect.Map:
		for _, f := range field.MapKeys() {
			if strings.HasPrefix(f.String(), prefix) {
				suggestions = append(suggestions, path+f.String())
			}
		}
	case reflect.Array, reflect.Slice:
		for f := 0; f < field.Len(); f++ {
			num := fmt.Sprintf("%d", f)
			if strings.HasPrefix(num, prefix) {
				suggestions = append(suggestions, path+num)
			}
		}
	default:
		return nil
	}

	// sort suggestions alphabetically
	slices.SortFunc(suggestions, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	return suggestions
}

func ReadField(conf any, fieldPath string) (any, error) {
	keySequence := strings.Split(fieldPath, ".")
	field := reflect.ValueOf(conf)
	for index, key := range keySequence {
		if index == 0 && len(key) == 0 {
			// edge case: first elem is empty
			continue
		}
		// if field is a pointer, dereference
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}
		t := field.Kind()
		switch t {
		case reflect.Struct:
			field = fieldByEncodingName(field, key)
			if !field.IsValid() {
				return nil, fmt.Errorf(INVALID_FIELD, key)
			}
		case reflect.Map:
			field = field.MapIndex(reflect.ValueOf(key))
			if !field.IsValid() {
				return nil, fmt.Errorf(INVALID_FIELD, key)
			}
		case reflect.Array, reflect.Slice:
			i, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("cannot index sequence field with non-numeric key '%s'", key)
			}
			if i < 0 || i >= field.Len() {
				return nil, fmt.Errorf("index %d out of range", i)
			}
			field = field.Index(i)
		default:
			return nil, fmt.Errorf("cannot address element of type '%s' with key '%s'", t.String(), key)
		}
	}
	return field.Interface(), nil
}

func (conf *Settings) WriteField(fieldPath, value string, append bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	writeValueHelper(reflect.ValueOf(conf), fieldPath, value, append)
	return nil
}

// recursive function to write a (nested) value inside a container; value will
// be parsed as yaml based on the type of the final element of the fieldpath. if
// append is true, value will be added to the Map or Slice final type; function
// will panic if not appending to proper type; if all goes well, the function
// will return the updated container
func writeValueHelper(container reflect.Value, fieldPath string, value string, append bool) reflect.Value {
	if container.Kind() == reflect.Ptr {
		container = container.Elem()
	}
	if len(fieldPath) != 0 {
		// walk property field path
		key, restpath, _ := strings.Cut(fieldPath, ".")
		switch container.Kind() {
		case reflect.Struct:
			field := fieldByEncodingName(container, key)
			if !field.IsValid() {
				panic(fmt.Errorf(INVALID_FIELD, key))
			}
			value := writeValueHelper(field, restpath, value, append)
			if field.Kind() == reflect.Ptr {
				value = value.Addr()
			}
			field.Set(value)

		case reflect.Map:
			field := container.MapIndex(reflect.ValueOf(key))
			if !field.IsValid() {
				panic(fmt.Errorf(INVALID_FIELD, key))
			}
			// make a clone
			newfield := reflect.New(field.Type())
			newfield.Elem().Set(field)
			// assign new field to map
			container.SetMapIndex(reflect.ValueOf(key), writeValueHelper(newfield, restpath, value, append))

		case reflect.Array, reflect.Slice:
			i, err := strconv.Atoi(key)
			if err != nil {
				panic(fmt.Errorf("cannot index sequence field with non-numeric key '%s'", key))
			}
			if i < 0 || i >= container.Len() {
				panic(fmt.Errorf("index %d out of range", i))
			}
			value := writeValueHelper(container.Index(i), restpath, value, append)
			if container.Index(i).Kind() == reflect.Ptr {
				value = value.Addr()
			}
			container.Index(i).Set(value)

		default:
			panic(fmt.Errorf("cannot address element of type '%s' with key '%s'", container.Kind().String(), key))
		}
	} else {
		// final value case:
		// new value will have the same type as the container
		newValtype := container.Type()
		if append {
			// except if we are appending to a slice, where type is slice's element type
			if container.Kind() == reflect.Slice {
				newValtype = container.Type().Elem()
			} else if container.Kind() != reflect.Map {
				// if we attempt to append to anything other than a slice or a map, fail
				panic(fmt.Errorf("cannot append value(s) to a '%s'", container.Kind().String()))
			}

			keyValuePair := regexKeyValue.FindStringSubmatch(value)
			if len(keyValuePair) >= 2 {
				value = fmt.Sprintf("{ %s: %s }", keyValuePair[1], keyValuePair[2])
			}
		}
		v := reflect.New(newValtype)
		// parse string as yaml inside the newly created value
		if err := ReadYamlInto(v.Interface(), strings.NewReader(value)); err != nil {
			panic(fmt.Errorf("'%s' is not a %s", value, container.Type().Kind().String()))
		}
		// when not appending ('set' case), new value will replace previous (container)
		if !append {
			return v.Elem()
		}
		// when appending ('add' case), handle map and slice differently
		if container.Kind() == reflect.Map {
			// edge case: appending to an empty (nil) container
			if container.IsNil() {
				container.Set(v.Elem())
			} else {
				iter := v.Elem().MapRange()
				for iter.Next() {
					container.SetMapIndex(iter.Key(), iter.Value())
				}
			}
		} else if container.Kind() == reflect.Slice {
			container = reflect.Append(container, v.Elem())
		}
	}
	// return final container to replace previous one
	return container
}

func ToDictionary(conf any) map[string]any {
	data, err := yaml.Marshal(conf)
	// to simplify return value, consume errors by panicking upon (unlikely) occurrence
	if err != nil {
		panic(err)
	}
	dict := map[string]any{}
	err = yaml.Unmarshal(data, &dict)
	if err != nil {
		panic(err)
	}
	return dict
}

func (conf *Settings) ReadConfiguration(configRoot string) error {
	return conf.ReadFromFile(filepath.Join(configRoot, ConfigFileName))
}

func (conf *Settings) ReadConfigurationRecursive() (string, error) {
	var cwd string
	var err error
	if cwd, err = os.Getwd(); err != nil {
		return "", err
	}
	for {
		err = conf.ReadConfiguration(cwd)

		// Found Configuration
		if err == nil {
			return cwd, nil
		}

		// Encountered error while parsing configuration
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		// if currDir ends in separator, it is the root
		if strings.HasSuffix(cwd, string(filepath.Separator)) {
			break
		} else {
			cwd = filepath.Dir(cwd)
		}
	}
	return "", nil
}

func (conf *Settings) WriteConfiguration(dir string, deltaBase *Settings) error {
	prev := ToDictionary(deltaBase)
	curr := ToDictionary(conf)
	delta := computeDelta(prev, curr)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	path := filepath.Join(dir, ConfigFileName)
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(INDENTATION)
	if err = enc.Encode(delta); err != nil {
		return errors.Join(err, out.Close())
	}

	return out.Close()
}

// get available templates
const TemplateInfoFile string = ".template-info.yaml"

type TemplateInfo struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Dependencies struct {
		Compile []string `yaml:"compile"`
		Runtime []string `yaml:"runtime"`
	} `yaml:"dependencies"`
	Path       string            `yaml:"-"`
	RegenFiles map[string]string `yaml:"regenFiles"`
	Base       string            `yaml:"base,omitempty"`
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

func GenerateTemplate(fs afero.Fs, info TemplateInfo, destFolder string, context any, templates map[string]TemplateInfo, visited []string, configTemplateField map[string]string) error {
	if visited == nil {
		visited = make([]string, 0, 2)
	}

	if info.Base != "" {
		for _, v := range visited {
			if v == info.Base {
				return fmt.Errorf("circular dependency detected: %s", info.Base)
			}
		}

		if baseTempl, exists := templates[info.Base]; exists {
			visited = append(visited, info.Base)
			err := GenerateTemplate(fs, baseTempl, destFolder, context, templates, visited, configTemplateField)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("base template not found: %s", info.Base)
		}
	}

	for templFile, destFile := range info.RegenFiles {
		templPath := filepath.Join(info.Path, templFile)

		if _, err := os.Stat(templPath); os.IsNotExist(err) {
			fmt.Printf("warning: template file '%s' does not exist in '%s' template folder\n", templFile, info.Name)
		}

		configTemplateField[templFile] = destFile
	}

	fileList, err := getFileList(fs, info.Path)
	if err != nil {
		return err
	}

	for _, path := range fileList {
		if filepath.Base(path) == TemplateInfoFile {
			continue
		} else if isRegenFile(path, info) {
			if _, err = fsutil.CopyFile(filepath.Join(info.Path, path), filepath.Join(destFolder, path)); err != nil {
				return err
			}
			continue
		}
		err := templating.RenderTemplateFile(fs, path, info.Path, destFolder, context)
		if err != nil {
			return err
		}
	}

	return nil
}

// helpers
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

func readTemplateInfo(fullPath string) (info TemplateInfo, err error) {
	yamlData, err := os.Open(fullPath)
	if err != nil {
		return TemplateInfo{}, err
	}
	defer yamlData.Close()

	if err = ReadYamlInto(&info, yamlData); err != nil {
		return
	}
	// TODO: validate template info
	info.Path = filepath.Dir(fullPath)
	return
}

func computeDelta(prev, curr map[string]any) map[string]any {
	delta := map[string]any{}
	for key, value := range curr {
		oldValue, found := prev[key]
		if found && reflect.DeepEqual(value, oldValue) {
			// values are identical (unchanged)
			continue
		} else if found && reflect.TypeOf(value) == reflect.TypeOf(oldValue) && reflect.TypeOf(value).Kind() == reflect.Map {
			// value exists in both and it is a dictionary
			delta[key] = computeDelta(oldValue.(map[string]any), value.(map[string]any))
		} else {
			// in all other cases, keep new value
			delta[key] = value
		}
	}
	return delta
}

// Find field by `encoding` tag name
func fieldByEncodingName(v reflect.Value, name string) reflect.Value {
	for _, field := range reflect.VisibleFields(v.Type()) {
		encodingName := getFieldName(field)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			continue
		} else {
			// match encoding tag name or field name if former is not present
			if encodingName == name {
				return v.FieldByIndex(field.Index)
			}
		}
	}

	return reflect.Value{}
}

func fieldNamesByEncodingPrefix(v reflect.Value, prefix, basePath string) []string {
	var fieldNames []string

	for _, field := range reflect.VisibleFields(v.Type()) {
		name := getFieldName(field)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			continue
		} else {
			if strings.HasPrefix(name, prefix) {
				fieldNames = append(fieldNames, basePath+name)
			}
		}
	}

	return fieldNames
}

func getFieldName(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup(ENCODING); ok {
		return strings.Split(tag, ",")[0]
	}
	return field.Name
}

func ReadYamlInto(value interface{}, in io.Reader) error {
	dec := yaml.NewDecoder(in)
	dec.KnownFields(STRICT_DECODING_OPT)
	return dec.Decode(value)
}

func isRegenFile(path string, info TemplateInfo) bool {
	_, exists := info.RegenFiles[filepath.ToSlash(path)]
	return exists
}
