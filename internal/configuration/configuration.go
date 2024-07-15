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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ENCODING            = "yaml"
	STRICT_DECODING_OPT = true
	INDENTATION         = 4
)

type BuildOptions struct {
	OutputType 	 string              `yaml:"outputType"`
	DebugMode      bool              `yaml:"debug"`
	SkipHostEnv    bool              `yaml:"skipHostEnv,omitempty"`
	Env            map[string]string `yaml:"env"`
}

type AppSettings struct {
	Lang         string            `yaml:"lang"`
	Title        string            `yaml:"title"`
	Name         string            `yaml:"name"`
	Author       string            `yaml:"author"`
	Description  string            `yaml:"description"`
	Version      string            `yaml:"version"`
	FQDN         string            `yaml:"fqdn"`
	DUID         string            `yaml:"duid"`
	Options      map[string]any    `yaml:"options"`
	Dependencies Dependencies      `yaml:"dependencies"`
	Env          map[string]string `yaml:"env"`
	Entrypoint   string            `yaml:"entrypoint"`
}

type Dependencies struct {
	Compile []string `yaml:"compile"`
	Runtime []string `yaml:"runtime"`
}

type ArchitectureSettings struct {
	Platform string `yaml:"platform"`
}

type ArchitecturesMap map[string]ArchitectureSettings

type ToolchainSettings struct {
	Image         string           `yaml:"image"`
	Architectures ArchitecturesMap `yaml:"architectures"`
}

type CrossCompileConfig struct {
	Enabled bool     `yaml:"enabled"`
	Image   string   `yaml:"image"`
	Args    []string `yaml:"args"`
}

type BuildSettings struct {
	Toolchains         ToolchainSettings  `yaml:"toolchains"`
	Default            string             `yaml:"default,omitempty"`
	Options            BuildOptions       `yaml:"options"`
	CrossCompile       CrossCompileConfig `yaml:"crossCompile"`
	DockerFileTemplate string             `yaml:"dockerFileTemplate,omitempty"`
}

type PublishMethod int

type AuthType int

const (
	PUBLISH_METHOD_UNDEFINED = iota
	PUBLISH_METHOD_LISTEN
	PUBLISH_METHOD_PUT
	PUBLISH_METHOD_COPY
)

const (
	publishMethodListenName = "listen"
	publishMethodPutName    = "put"
	publishMethodCopyName   = "copy"
)

const (
	AUTH_UNDEFINED = iota
	AUTH_HTTP_BASIC
	AUTH_HTTP_BEARER
	AUTH_HTTP_DIGEST
	AUTH_SSH_PASSWORD
	AUTH_SSH_PUBLIC_KEY
)

const (
	ConfigFileName = "corteca.yaml"
)

const (
	authHttpBasicName    = "basic"
	authHttpBearerName   = "bearer"
	authHttpDigestName   = "digest"
	authSshPasswordName  = "password"
	authSshPublicKeyName = "publicKey"
)

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
	default:
		out = nil
		err = fmt.Errorf("invalid publish method (%v)", m)
	}

	return strings.TrimSpace(string(out)), err

}

func (a AuthType) MarshalYAML() (interface{}, error) {
	var out []byte
	var err error

	switch a {
	case AUTH_HTTP_BASIC:
		out, err = yaml.Marshal(authHttpBasicName)
	case AUTH_HTTP_BEARER:
		out, err = yaml.Marshal(authHttpBearerName)
	case AUTH_HTTP_DIGEST:
		out, err = yaml.Marshal(authHttpDigestName)
	case AUTH_SSH_PASSWORD:
		out, err = yaml.Marshal(authSshPasswordName)
	case AUTH_SSH_PUBLIC_KEY:
		out, err = yaml.Marshal(authSshPublicKeyName)
	default:
		out = nil
		err = fmt.Errorf("invalid authorization type (%v)", a)
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
	default:
		return fmt.Errorf("unrecognized publish method '%v'", name)
	}
	return nil
}

func (a *AuthType) UnmarshalYAML(data *yaml.Node) error {
	var name string
	if err := yaml.Unmarshal([]byte(data.Value), &name); err != nil {
		return err
	}

	switch name {
	case authHttpBasicName:
		*a = AUTH_HTTP_BASIC
	case authHttpBearerName:
		*a = AUTH_HTTP_BEARER
	case authHttpDigestName:
		*a = AUTH_HTTP_DIGEST
	case authSshPasswordName:
		*a = AUTH_SSH_PASSWORD
	case authSshPublicKeyName:
		*a = AUTH_SSH_PUBLIC_KEY
	default:
		return fmt.Errorf("unrecognized authorization type '%v'", name)
	}
	return nil
}

type Endpoint struct {
	Addr           string   `yaml:"addr,omitempty"`
	Auth           AuthType `yaml:"auth,omitempty"`
	PrivateKeyFile string   `yaml:"privateKeyFile,omitempty"`
	Token          string   `yaml:"token,omitempty"`
}

type PublishTarget struct {
	Endpoint `yaml:",omitempty,inline"`
	Method   PublishMethod `yaml:"method,omitempty"`
}

type SequenceCmd struct {
	Cmd     string `yaml:"cmd,omitempty"`
	Output  string `yaml:"expectedOutput,omitempty"`
	Delay   uint   `yaml:"delay,omitempty"`
	Retries uint   `yaml:"retries,omitempty"`
}

type DownloadSource struct {
	Url     string `yaml:"url,omitempty"`
	Publish string `yaml:"publish,omitempty"`
}

type DeployDevice struct {
	Endpoint `yaml:",omitempty,inline"`
	Source   DownloadSource `yaml:"source,omitempty"`
}

type DeploySettings struct {
	Sequence []SequenceCmd `yaml:"sequence,omitempty"`
	LogFile  string        `yaml:"logFile,omitempty"`
}

type Settings struct {
	App     AppSettings              `yaml:"app"`
	Build   BuildSettings            `yaml:"build"`
	Deploy  DeploySettings           `yaml:"deploy"`
	Publish map[string]PublishTarget `yaml:"publish,omitempty"`
	Devices map[string]DeployDevice  `yaml:"devices,omitempty"`
}

func NewConfiguration() Settings {
	return Settings{
		App: AppSettings{
			Options: map[string]any{},
		},
		Build: BuildSettings{
			Toolchains: ToolchainSettings{
				Image:         "",
				Architectures: make(ArchitecturesMap),
			},
		},
		Deploy: DeploySettings{
			Sequence: []SequenceCmd{},
		},
		Publish: map[string]PublishTarget{},
		Devices: map[string]DeployDevice{},
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

func (conf Settings) ReadField(fieldPath string) (any, error) {
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
				return nil, fmt.Errorf("invalid field '%s'", key)
			}
		case reflect.Map:
			field = field.MapIndex(reflect.ValueOf(key))
			if !field.IsValid() {
				return nil, fmt.Errorf("invalid field '%s'", key)
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
				panic(fmt.Errorf("invalid field '%s'", key))
			}
			field.Set(writeValueHelper(field, restpath, value, append))

		case reflect.Map:
			field := container.MapIndex(reflect.ValueOf(key))
			if !field.IsValid() {
				panic(fmt.Errorf("invalid field '%s'", key))
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
			container.Index(i).Set(writeValueHelper(container.Index(i), restpath, value, append))

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
		}
		v := reflect.New(newValtype)
		// accept special format of KEY=VALUE (for shell convenience) and reformat it as yaml
		if key, val, found := strings.Cut(value, "="); found {
			value = fmt.Sprintf("{ %s: %s }", key, val)
		}
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

// helpers

func computeDelta(prev, curr map[string]any) map[string]any {
	delta := map[string]any{}
	for key, value := range curr {
		oldValue, found := prev[key]
		if found && reflect.DeepEqual(value, oldValue) {
			// values are identical (unchanged)
			continue
		} else if found && reflect.TypeOf(value).Kind() == reflect.Map && reflect.TypeOf(oldValue).Kind() == reflect.Map {
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
	for i := 0; i < v.Type().NumField(); i++ {
		tag := v.Type().Field(i).Tag

		if encodingTag, ok := tag.Lookup(ENCODING); ok {
			// `encoding` tag (ex: json, yaml, xml) consists comma-separated values, where first one is always the encoding's name
			encodingName := strings.Split(encodingTag, ",")[0]
			// match encoding tag name, or field name if former is not present
			if encodingName == name {
				return v.Field(i)
			}
		}
	}
	// field name not found, search anonymous (embedded) structs
	for i := 0; i < v.Type().NumField(); i++ {
		if v.Type().Field(i).Anonymous && v.Type().Field(i).Type.Kind() == reflect.Struct {
			value := fieldByEncodingName(v.Field(i), name)
			if value.IsValid() {
				return value
			}
		}
	}
	return reflect.Value{}
}

func ReadYamlInto(value interface{}, in io.Reader) error {
	dec := yaml.NewDecoder(in)
	dec.KnownFields(STRICT_DECODING_OPT)
	return dec.Decode(value)
}
