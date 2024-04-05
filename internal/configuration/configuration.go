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
	"math/bits"
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
	SkipHostEnv bool              `yaml:"skipHostEnv,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
}

type AppSettings struct {
	Lang        string         `yaml:"lang,omitempty"`
	Title       string         `yaml:"title,omitempty"`
	Name        string         `yaml:"name,omitempty"`
	Author      string         `yaml:"author,omitempty"`
	Description string         `yaml:"description,omitempty"`
	Version     string         `yaml:"version,omitempty"`
	FQDN        string         `yaml:"fqdn,omitempty"`
	DUID        string         `yaml:"duid,omitempty"`
	Options     map[string]any `yaml:"options,omitempty"`
}

type ToolchainSettings struct {
	Image      string `yaml:"image,omitempty"`
	ConfigFile string `yaml:"configFile,omitempty"`
}

type Toolchains map[string]ToolchainSettings

type BuildSettings struct {
	Toolchains Toolchains   `yaml:"toolchains,omitempty"`
	Default    string       `yaml:"default,omitempty"`
	Options    BuildOptions `yaml:"options,omitempty"`
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
	App     AppSettings              `yaml:"app,omitempty"`
	Build   BuildSettings            `yaml:"build,omitempty"`
	Deploy  DeploySettings           `yaml:"deploy,omitempty"`
	Publish map[string]PublishTarget `yaml:"publish,omitempty"`
	Devices map[string]DeployDevice  `yaml:"devices,omitempty"`
}

func (toolchain *Toolchains) UnmarshalYAML(data *yaml.Node) error {
	if data.Kind != yaml.MappingNode {
		return errors.New("wrong value type")
	}

	for i := 0; i < len(data.Content); i += 2 {
		var alias string
		if err := data.Content[i].Decode(&alias); err != nil {
			return err
		}
		settings := (*toolchain)[alias]
		if err := data.Content[i+1].Decode(&settings); err != nil {
			return err
		}
		(*toolchain)[alias] = settings
	}

	return nil
}

func NewConfiguration() Settings {
	return Settings{
		App: AppSettings{
			Options: map[string]any{},
		},
		Build: BuildSettings{
			Toolchains: map[string]ToolchainSettings{},
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

	decoder := yaml.NewDecoder(in)
	decoder.KnownFields(STRICT_DECODING_OPT)

	if err = decoder.Decode(conf); err != nil {
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

func (conf *Settings) ReadField(fieldPath string, output io.Writer) error {
	field, _, err := retrieveField(reflect.ValueOf(conf), fieldPath, false)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(output).Encode(field.Interface())
}

func (conf *Settings) WriteField(fieldPath, value string) error {
	field, restPath, err := retrieveField(reflect.ValueOf(conf), fieldPath, true)
	if err != nil {
		return err
	}
	if restPath != "" {
		m := field.Elem().Interface().(map[string]any)
		setKeyValuePair(m, restPath, value)
	} else if !field.CanSet() {
		return fmt.Errorf("cannot set value of '%v'", fieldPath)
	} else {
		switch field.Type().Kind() {
		case reflect.Bool:
			v, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.SetBool(v)
		case reflect.Uint:
			v, err := strconv.ParseUint(value, 10, bits.UintSize)
			if err != nil {
				return err
			}
			field.SetUint(v)
		case reflect.Int:
			v, err := strconv.ParseInt(value, 10, bits.UintSize)
			if err != nil {
				return err
			}
			field.SetInt(v)
		case reflect.String:
			field.SetString(value)
		default:
			return fmt.Errorf("unsupported field type '%v'", field.Kind().String())
		}
	}
	return nil
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

func retrieveField(v reflect.Value, fieldPath string, addressable bool) (reflect.Value, string, error) {
	for {
		// if we need addressable values inside a map (with string keys) we need
		// to leave reflection and provide access to the map directly
		if addressable && v.Elem().Type().Kind() == reflect.Map && v.Elem().Type().Key().Kind() == reflect.String {
			return v, fieldPath, nil
		}
		fieldName, restPath, _ := strings.Cut(fieldPath, ".")
		field := fieldByEncodingName(v.Elem(), fieldName, ENCODING)
		if !field.IsValid() {
			return reflect.Value{}, restPath, fmt.Errorf("field '%v' not found", fieldName)
		} else if restPath == "" {
			return field, "", nil
		}
		v = field.Addr()
		fieldPath = restPath
	}
}

// Find field by `encoding` tag name
func fieldByEncodingName(v reflect.Value, name string, encoding string) reflect.Value {
	if v.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	for i := 0; i < v.Type().NumField(); i++ {
		tag := v.Type().Field(i).Tag

		if encodingTag, ok := tag.Lookup(encoding); ok {
			// `encoding` tag (ex: json, yaml, xml) consists comma-separated values, where first one is always the encoding's name
			EncodingName := strings.Split(encodingTag, ",")[0]
			if EncodingName == name {
				return v.Field(i)
			}
		}
	}
	return reflect.Value{}
}

func setKeyValuePair(v map[string]any, fieldPath string, value any) error {
	for {
		fieldName, restPath, _ := strings.Cut(fieldPath, ".")
		if restPath == "" {
			v[fieldName] = value
			return nil
		} else {
			v[fieldName] = map[string]any{}
			v = v[fieldName].(map[string]any)
			fieldPath = restPath
		}
	}
}

